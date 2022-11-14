package pbapi

import (
	"fmt"
	"log"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/resolvers"
	"github.com/pocketbase/pocketbase/tools/search"
	"github.com/spf13/cast"
)

func RecordView(app core.App, collectionNameOrId, queryString, expand string) ([]*models.Record, error) {
	collection, err := app.Dao().FindCollectionByNameOrId(collectionNameOrId)
	if err != nil {
		return nil, err
	}

	requestData := map[string]any{}

	fieldsResolver := resolvers.NewRecordFieldResolver(
		app.Dao(),
		collection,
		requestData,
		// hidden fields are searchable only by admins
		false,
	)

	searchProvider := search.NewProvider(fieldsResolver).
		Query(app.Dao().RecordQuery(collection))

	var rawRecords = []dbx.NullStringMap{}
	var result *search.Result
	result, err = searchProvider.ParseAndExec(queryString, &rawRecords)
	if err != nil {
		return nil, err
	}

	records := models.NewRecordsFromNullStringMaps(collection, rawRecords)

	// expand records relations
	expands := strings.Split(expand, ",")
	if len(expands) > 0 {
		failed := app.Dao().ExpandRecords(
			records,
			expands,
			expandFetch(app.Dao(), false, requestData),
		)
		if len(failed) > 0 && app.IsDebug() {
			log.Println("Failed to expand relations: ", failed)
		}
	}
	result.Items = records
	return records, nil
}

func RecordList(app core.App, collectionNameOrId, queryString, expand string) (*search.Result, error) {
	collection, err := app.Dao().FindCollectionByNameOrId(collectionNameOrId)
	if err != nil {
		return nil, err
	}

	requestData := map[string]any{}

	fieldsResolver := resolvers.NewRecordFieldResolver(
		app.Dao(),
		collection,
		requestData,
		// hidden fields are searchable only by admins
		false,
	)

	searchProvider := search.NewProvider(fieldsResolver).
		Query(app.Dao().RecordQuery(collection))

	var rawRecords = []dbx.NullStringMap{}
	var result *search.Result
	result, err = searchProvider.ParseAndExec(queryString, &rawRecords)
	if err != nil {
		return nil, err
	}

	records := models.NewRecordsFromNullStringMaps(collection, rawRecords)

	// expand records relations
	expands := strings.Split(expand, ",")
	if len(expands) > 0 {
		failed := app.Dao().ExpandRecords(
			records,
			expands,
			expandFetch(app.Dao(), false, requestData),
		)
		if len(failed) > 0 && app.IsDebug() {
			log.Println("Failed to expand relations: ", failed)
		}
	}
	result.Items = records
	return result, nil
}

// expandFetch is the records fetch function that is used to expand related records.
func expandFetch(
	dao *daos.Dao,
	isAdmin bool,
	requestData map[string]any,
) daos.ExpandFetchFunc {
	return func(relCollection *models.Collection, relIds []string) ([]*models.Record, error) {
		records, err := dao.FindRecordsByIds(relCollection.Id, relIds, func(q *dbx.SelectQuery) error {
			if isAdmin {
				return nil // admins can access everything
			}

			if relCollection.ViewRule == nil {
				return fmt.Errorf("Only admins can view collection %q records", relCollection.Name)
			}

			if *relCollection.ViewRule != "" {
				resolver := resolvers.NewRecordFieldResolver(dao, relCollection, requestData, true)
				expr, err := search.FilterData(*(relCollection.ViewRule)).BuildExpr(resolver)
				if err != nil {
					return err
				}
				resolver.UpdateQuery(q)
				q.AndWhere(expr)
			}

			return nil
		})

		return records, err
	}
}

// Create - create single record in collection with provided fields/values (fieldsMap)
// check rights/rules
// don't trigger hooks/logs event
// return error or record struct
func RecordCreate(app core.App, collectionNameOrId string, admin *models.Admin, fieldsMap map[string]any) (*models.Record, error) {
	collection, err := app.Dao().FindCollectionByNameOrId(collectionNameOrId)
	if err != nil {
		return nil, err
	}

	hasFullManageAccess, errCreate := createTest(app, collection, admin, fieldsMap)
	if errCreate != nil {
		return nil, errCreate
	}

	record := models.NewRecord(collection)
	form := forms.NewRecordUpsert(app, record)
	form.SetFullManageAccess(hasFullManageAccess)

	if err := form.LoadData(fieldsMap); err != nil {
		return nil, err
	}
	// create the record
	if err := form.Submit(); err != nil {
		return nil, err
	}
	return record, nil
}

// createTest check access and create rules before create if admin is nil
func createTest(app core.App, collection *models.Collection, admin *models.Admin, fieldsMap map[string]any) (bool, error) {
	hasFullManageAccess := admin != nil

	if collection == nil {
		return hasFullManageAccess, apis.NewNotFoundError("", "Missing collection context.")
	}
	if admin == nil && collection.CreateRule == nil {
		// only admins can access if the rule is nil
		return hasFullManageAccess, apis.NewForbiddenError("Only admins can perform this action.", nil)
	}

	// temporary save the record and check it against the create rule
	if admin == nil && collection.CreateRule != nil {
		createRuleFunc := func(q *dbx.SelectQuery) error {
			if *collection.CreateRule == "" {
				return nil // no create rule to resolve
			}

			resolver := resolvers.NewRecordFieldResolver(app.Dao(), collection, fieldsMap, true)
			expr, err := search.FilterData(*collection.CreateRule).BuildExpr(resolver)
			if err != nil {
				return err
			}
			resolver.UpdateQuery(q)
			q.AndWhere(expr)
			return nil
		}

		testRecord := models.NewRecord(collection)
		testForm := forms.NewRecordUpsert(app, testRecord)
		testForm.SetFullManageAccess(true)
		if err := testForm.LoadData(fieldsMap); err != nil {
			return hasFullManageAccess, apis.NewBadRequestError("Failed to load the submitted data due to invalid formatting.", err)
		}

		testErr := testForm.DrySubmit(func(txDao *daos.Dao) error {
			foundRecord, err := txDao.FindRecordById(collection.Id, testRecord.Id, createRuleFunc)
			if err != nil {
				return fmt.Errorf("DrySubmit create rule failure: %v", err)
			}
			hasFullManageAccess = hasAuthManageAccess(txDao, foundRecord, fieldsMap)
			return nil
		})

		if testErr != nil {
			return hasFullManageAccess, apis.NewBadRequestError("Failed to create record.", testErr)
		}
	}
	return hasFullManageAccess, nil
}

// hasAuthManageAccess checks whether the client is allowed to have full
// [forms.RecordUpsert] auth management permissions
// (aka. allowing to change system auth fields without oldPassword).
func hasAuthManageAccess(
	dao *daos.Dao,
	record *models.Record,
	requestData map[string]any,
) bool {
	if !record.Collection().IsAuth() {
		return false
	}

	manageRule := record.Collection().AuthOptions().ManageRule

	if manageRule == nil || *manageRule == "" {
		return false // only for admins (manageRule can't be empty)
	}

	if auth, ok := requestData["auth"].(map[string]any); !ok || cast.ToString(auth["id"]) == "" {
		return false // no auth record
	}

	ruleFunc := func(q *dbx.SelectQuery) error {
		resolver := resolvers.NewRecordFieldResolver(dao, record.Collection(), requestData, true)
		expr, err := search.FilterData(*manageRule).BuildExpr(resolver)
		if err != nil {
			return err
		}
		resolver.UpdateQuery(q)
		q.AndWhere(expr)
		return nil
	}

	_, findErr := dao.FindRecordById(record.Collection().Id, record.Id, ruleFunc)

	return findErr == nil
}
