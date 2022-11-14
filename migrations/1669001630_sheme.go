package migrations

import (
	"fmt"
	"log"

	"github.com/pocketbase/dbx"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(db dbx.Builder) error {
		log.Println("vsd_migrate started")
		if _, err := db.AddColumn("users", "photo_url", `TEXT`).Execute(); err != nil {
			return err
		}
		migrateQuery := db.NewQuery(
			`
				delete from '_collections';
				
				CREATE TABLE 'domain' ('created' TEXT DEFAULT '' NOT NULL, 'descr' TEXT DEFAULT '', 'hostname' TEXT DEFAULT '', 'icon' TEXT DEFAULT '', 'id' TEXT PRIMARY KEY, 'lang' TEXT DEFAULT '', 'title' TEXT DEFAULT '', 'updated' TEXT DEFAULT '' NOT NULL, 'url' TEXT DEFAULT '');
				CREATE INDEX '_51lodztkxakfboh_created_idx' ON 'domain' ('created');
				CREATE TABLE 'feed' ('created' TEXT DEFAULT '' NOT NULL, 'descr' TEXT DEFAULT '', 'domain_id' TEXT DEFAULT '', 'id' TEXT PRIMARY KEY, 'last_error' TEXT DEFAULT '', 'last_fetch' TEXT DEFAULT '', 'resp_code' REAL DEFAULT 0, 'title' TEXT DEFAULT '', 'updated' TEXT DEFAULT '' NOT NULL, 'url' TEXT DEFAULT '');
				CREATE INDEX '_to82imph7oc7bqi_created_idx' ON 'feed' ('created');
				CREATE TABLE 'post' ('created' TEXT DEFAULT '' NOT NULL, 'descr' TEXT DEFAULT '', 'feed_id' TEXT DEFAULT '', 'id' TEXT PRIMARY KEY, 'img' TEXT DEFAULT '', 'pub_date' TEXT DEFAULT '', 'sum_html' TEXT DEFAULT '', 'sum_txt' TEXT DEFAULT '', 'title' TEXT DEFAULT '', 'updated' TEXT DEFAULT '' NOT NULL, 'url' TEXT DEFAULT '');
				CREATE INDEX '_d6mbq3heomws7j8_created_idx' ON 'post' ('created');
				CREATE TABLE 'usr_feed' ('created' TEXT DEFAULT '' NOT NULL, 'feed_id' TEXT DEFAULT '', 'id' TEXT PRIMARY KEY, 'updated' TEXT DEFAULT '' NOT NULL, 'user_id' TEXT DEFAULT '');
				CREATE UNIQUE INDEX _usr_feed_idx on 'usr_feed' ('feed_id', 'user_id');
				CREATE INDEX '_hgdnbtp4stpg8j2_created_idx' ON 'usr_feed' ('created');
				
				INSERT INTO _collections VALUES('51lodztkxakfboh',0,'base','domain','[{"system":false,"id":"hb9kuxjn","name":"url","type":"url","required":false,"unique":true,"options":{"exceptDomains":null,"onlyDomains":null}},{"system":false,"id":"n4cx64x1","name":"title","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"bbbxair0","name":"descr","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"7uzxmyv7","name":"icon","type":"url","required":false,"unique":false,"options":{"exceptDomains":null,"onlyDomains":null}},{"system":false,"id":"oippqcym","name":"lang","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"e7lz2se0","name":"hostname","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}}]','','',NULL,NULL,NULL,'{}','2022-11-10 15:02:01.864Z','2022-11-10 15:02:01.864Z');
				INSERT INTO _collections VALUES('to82imph7oc7bqi',0,'base','feed','[{"system":false,"id":"bvk26iyf","name":"domain_id","type":"relation","required":true,"unique":false,"options":{"maxSelect":1,"collectionId":"51lodztkxakfboh","cascadeDelete":true}},{"system":false,"id":"l98bchxu","name":"url","type":"url","required":false,"unique":true,"options":{"exceptDomains":[],"onlyDomains":[]}},{"system":false,"id":"ohf4fngr","name":"title","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"kpbnpjgx","name":"descr","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"iduiheeu","name":"last_fetch","type":"date","required":false,"unique":false,"options":{"min":"","max":""}},{"system":false,"id":"y55wbwgm","name":"last_error","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"lknrpo8c","name":"resp_code","type":"number","required":false,"unique":false,"options":{"min":null,"max":null}}]','','',NULL,NULL,NULL,'{}','2022-11-10 15:02:01.864Z','2022-11-10 15:02:01.864Z');
				INSERT INTO _collections VALUES('d6mbq3heomws7j8',0,'base','post','[{"system":false,"id":"kyz9t78c","name":"feed_id","type":"relation","required":true,"unique":false,"options":{"maxSelect":1,"collectionId":"to82imph7oc7bqi","cascadeDelete":true}},{"system":false,"id":"6n8vvbod","name":"url","type":"url","required":false,"unique":true,"options":{"exceptDomains":null,"onlyDomains":null}},{"system":false,"id":"jcdaujko","name":"title","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"6odm08le","name":"descr","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"30qjjfju","name":"img","type":"url","required":false,"unique":false,"options":{"exceptDomains":null,"onlyDomains":null}},{"system":false,"id":"9hpiahga","name":"sum_html","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"ylir5uwu","name":"sum_txt","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"btktqs4t","name":"pub_date","type":"date","required":false,"unique":false,"options":{"min":"","max":""}}]',NULL,NULL,NULL,NULL,NULL,'{}','2022-11-10 15:02:01.865Z','2022-11-10 15:02:01.865Z');
				INSERT INTO _collections VALUES('hgdnbtp4stpg8j2',0,'base','usr_feed','[{"system":false,"id":"lqqklaff","name":"user_id","type":"relation","required":false,"unique":false,"options":{"maxSelect":1,"collectionId":"users","cascadeDelete":true}},{"system":false,"id":"rldjtobp","name":"feed_id","type":"relation","required":false,"unique":false,"options":{"maxSelect":1,"collectionId":"to82imph7oc7bqi","cascadeDelete":true}}]','',NULL,'',NULL,NULL,'{}','2022-11-10 15:02:01.866Z','2022-11-10 15:02:01.866Z');
				INSERT INTO _collections VALUES('2ykc8bz1hznpsdg',0,'auth','users','[{"system":false,"id":"adcw8cbp","name":"photo_url","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}}]',NULL,NULL,NULL,NULL,NULL,'{"allowEmailAuth":true,"allowOAuth2Auth":true,"allowUsernameAuth":true,"exceptEmailDomains":null,"manageRule":null,"minPasswordLength":8,"onlyEmailDomains":null,"requireEmail":false}','2022-11-10 17:41:28.045Z','2022-11-10 17:41:28.045Z');
				`)

		res, errMigrate := migrateQuery.Execute()

		fmt.Println("migrate", res, errMigrate)
		return errMigrate
	}, func(db dbx.Builder) error {
		// add down queries...

		return nil
	})
}
