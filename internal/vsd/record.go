package vsd

import (
	"fmt"
	"log"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/daos"
	"github.com/pocketbase/pocketbase/forms"
	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/resolvers"
	"github.com/pocketbase/pocketbase/tools/search"
)

func CreateRecord(app core.App, collectionNameOrId string, requestData map[string]any) (*models.Record, error) {
	collection, err := app.Dao().FindCollectionByNameOrId(collectionNameOrId)
	if err != nil {
		return nil, err
	}
	record := models.NewRecord(collection)
	form := forms.NewRecordUpsert(app, record)

	if err := form.LoadData(requestData); err != nil {
		return nil, err
	}
	// create the record
	if err := form.Submit(); err != nil {
		return nil, err
	}
	return record, err
}

func RecordsList(app core.App, collectionNameOrId, queryString, expand string) (*search.Result, error) {
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
