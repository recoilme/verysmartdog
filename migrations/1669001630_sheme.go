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
		if _, err := db.AddColumn("users", "params", `TEXT`).Execute(); err != nil {
			return err
		}
		migrateQuery := db.NewQuery(
			`
				delete from '_collections';
				
				CREATE TABLE IF NOT EXISTS 'domain' ('created' TEXT DEFAULT '' NOT NULL, 'descr' TEXT DEFAULT '', 'hostname' TEXT DEFAULT '', 'icon' TEXT DEFAULT '', 'id' TEXT PRIMARY KEY, 'lang' TEXT DEFAULT '', 'title' TEXT DEFAULT '', 'updated' TEXT DEFAULT '' NOT NULL, 'url' TEXT DEFAULT '', "params" TEXT DEFAULT '');
				CREATE TABLE IF NOT EXISTS 'feed' ('created' TEXT DEFAULT '' NOT NULL, 'descr' TEXT DEFAULT '', 'domain_id' TEXT DEFAULT '', 'id' TEXT PRIMARY KEY, 'last_error' TEXT DEFAULT '', 'last_fetch' TEXT DEFAULT '', 'resp_code' REAL DEFAULT 0, 'title' TEXT DEFAULT '', 'updated' TEXT DEFAULT '' NOT NULL, 'url' TEXT DEFAULT '', "lang" TEXT DEFAULT '', "icon" TEXT DEFAULT '', "pub_date" TEXT DEFAULT '', "params" TEXT DEFAULT '');
				CREATE TABLE IF NOT EXISTS 'post' ('created' TEXT DEFAULT '' NOT NULL, 'descr' TEXT DEFAULT '', 'feed_id' TEXT DEFAULT '', 'id' TEXT PRIMARY KEY, 'img' TEXT DEFAULT '', 'pub_date' TEXT DEFAULT '', 'sum_html' TEXT DEFAULT '', 'sum_txt' TEXT DEFAULT '', 'title' TEXT DEFAULT '', 'updated' TEXT DEFAULT '' NOT NULL, 'url' TEXT DEFAULT '', "author" TEXT DEFAULT '', "category" TEXT DEFAULT '', "guid" TEXT DEFAULT '', "params" TEXT DEFAULT '');
				CREATE TABLE IF NOT EXISTS 'usr_feed' ('created' TEXT DEFAULT '' NOT NULL, 'feed_id' TEXT DEFAULT '', 'id' TEXT PRIMARY KEY, 'updated' TEXT DEFAULT '' NOT NULL, 'user_id' TEXT DEFAULT '', "order" REAL DEFAULT 0, "params" TEXT DEFAULT '');
				CREATE TABLE IF NOT EXISTS 'keyword' ('created' TEXT DEFAULT '' NOT NULL, 'id' TEXT PRIMARY KEY, 'idf' REAL DEFAULT 0, 'type' REAL DEFAULT 0, 'updated' TEXT DEFAULT '' NOT NULL, "keyword" TEXT DEFAULT '');
				CREATE TABLE IF NOT EXISTS 'post_keyword' ('created' TEXT DEFAULT '' NOT NULL, 'id' TEXT PRIMARY KEY, 'keyword_id' TEXT DEFAULT '', 'post_id' TEXT DEFAULT '', 'updated' TEXT DEFAULT '' NOT NULL);

				CREATE INDEX '_domain_created_idx' ON 'domain' ('created');
				CREATE INDEX '_feed_created_idx' ON 'feed' ('created');
				CREATE INDEX 'feed_last_fetch_idx' ON 'feed' ('last_fetch');
				CREATE INDEX '_post_created_idx' ON 'post' ('created');
				CREATE INDEX '_post_guid_idx' ON 'post' ('guid');
				CREATE UNIQUE INDEX _usr_feed_idx on 'usr_feed' ('feed_id', 'user_id');
				CREATE INDEX '_usr_feed_created_idx' ON 'usr_feed' ('created');
				CREATE INDEX '_keyword_created_idx' ON 'keyword' ('created');
				CREATE UNIQUE INDEX _post_keyword_idx on 'post_keyword' ('keyword_id', 'post_id');
				CREATE INDEX '_post_keyword_created_idx' ON 'post_keyword' ('created');
				
				INSERT INTO _collections VALUES('51lodztkxakfboh',0,'base','domain','[{"system":false,"id":"hb9kuxjn","name":"url","type":"url","required":false,"unique":true,"options":{"exceptDomains":null,"onlyDomains":null}},{"system":false,"id":"n4cx64x1","name":"title","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"bbbxair0","name":"descr","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"7uzxmyv7","name":"icon","type":"url","required":false,"unique":false,"options":{"exceptDomains":null,"onlyDomains":null}},{"system":false,"id":"oippqcym","name":"lang","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"e7lz2se0","name":"hostname","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"dk6oftta","name":"params","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}}]',NULL,'',NULL,NULL,NULL,'{}','2022-11-10 15:02:01.864Z','2022-12-01 13:35:14.471Z');
				INSERT INTO _collections VALUES('to82imph7oc7bqi',0,'base','feed','[{"system":false,"id":"bvk26iyf","name":"domain_id","type":"relation","required":true,"unique":false,"options":{"maxSelect":1,"collectionId":"51lodztkxakfboh","cascadeDelete":true}},{"system":false,"id":"l98bchxu","name":"url","type":"url","required":false,"unique":true,"options":{"exceptDomains":[],"onlyDomains":[]}},{"system":false,"id":"ohf4fngr","name":"title","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"kpbnpjgx","name":"descr","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"iduiheeu","name":"last_fetch","type":"date","required":false,"unique":false,"options":{"min":"","max":""}},{"system":false,"id":"y55wbwgm","name":"last_error","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"lknrpo8c","name":"resp_code","type":"number","required":false,"unique":false,"options":{"min":null,"max":null}},{"system":false,"id":"tadncnp2","name":"lang","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"k6rcrjgz","name":"icon","type":"url","required":false,"unique":false,"options":{"exceptDomains":null,"onlyDomains":null}},{"system":false,"id":"e20t0mjs","name":"pub_date","type":"date","required":false,"unique":false,"options":{"min":"","max":""}},{"system":false,"id":"htyilgou","name":"params","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}}]',NULL,'',NULL,NULL,NULL,'{}','2022-11-10 15:02:01.864Z','2022-12-01 13:34:39.077Z');
				INSERT INTO _collections VALUES('d6mbq3heomws7j8',0,'base','post','[{"system":false,"id":"kyz9t78c","name":"feed_id","type":"relation","required":true,"unique":false,"options":{"maxSelect":1,"collectionId":"to82imph7oc7bqi","cascadeDelete":true}},{"system":false,"id":"6n8vvbod","name":"url","type":"url","required":false,"unique":true,"options":{"exceptDomains":null,"onlyDomains":null}},{"system":false,"id":"jcdaujko","name":"title","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"6odm08le","name":"descr","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"30qjjfju","name":"img","type":"url","required":false,"unique":false,"options":{"exceptDomains":null,"onlyDomains":null}},{"system":false,"id":"9hpiahga","name":"sum_html","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"ylir5uwu","name":"sum_txt","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"btktqs4t","name":"pub_date","type":"date","required":false,"unique":false,"options":{"min":"","max":""}},{"system":false,"id":"nipxtyc8","name":"author","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"1vlz8zoc","name":"category","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"cryi40a8","name":"guid","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"lom5cvz5","name":"params","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}}]',NULL,'',NULL,NULL,NULL,'{}','2022-11-10 15:02:01.865Z','2022-12-01 13:43:24.227Z');
				INSERT INTO _collections VALUES('hgdnbtp4stpg8j2',0,'base','usr_feed','[{"system":false,"id":"lqqklaff","name":"user_id","type":"relation","required":false,"unique":false,"options":{"maxSelect":1,"collectionId":"users","cascadeDelete":true}},{"system":false,"id":"rldjtobp","name":"feed_id","type":"relation","required":false,"unique":false,"options":{"maxSelect":1,"collectionId":"to82imph7oc7bqi","cascadeDelete":true}},{"system":false,"id":"yrvn9qbu","name":"order","type":"number","required":false,"unique":false,"options":{"min":null,"max":null}},{"system":false,"id":"5huytven","name":"params","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}}]',NULL,'','',NULL,NULL,'{}','2022-11-10 15:02:01.866Z','2022-12-01 13:43:36.988Z');
				INSERT INTO _collections VALUES('2ykc8bz1hznpsdg',0,'auth','users','[{"system":false,"id":"adcw8cbp","name":"photo_url","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"becpywvb","name":"params","type":"text","required":false,"unique":false,"options":{"min":null,"max":null,"pattern":""}}]',NULL,NULL,NULL,NULL,NULL,'{"allowEmailAuth":true,"allowOAuth2Auth":true,"allowUsernameAuth":true,"exceptEmailDomains":null,"manageRule":null,"minPasswordLength":8,"onlyEmailDomains":null,"requireEmail":false}','2022-11-10 17:41:28.045Z','2022-11-25 10:03:15.803Z');
				INSERT INTO _collections VALUES('ue9kdud987b9z5z',0,'base','keyword','[{"system":false,"id":"unk2ooie","name":"keyword","type":"text","required":false,"unique":true,"options":{"min":null,"max":null,"pattern":""}},{"system":false,"id":"5nq8v2gx","name":"type","type":"number","required":false,"unique":false,"options":{"min":null,"max":null}},{"system":false,"id":"7lwvnwyn","name":"idf","type":"number","required":false,"unique":false,"options":{"min":null,"max":null}}]',NULL,'',NULL,NULL,NULL,'{}','2022-11-27 09:20:48.277Z','2022-12-01 13:44:05.880Z');
				INSERT INTO _collections VALUES('dz3w892eyxmv9th',0,'base','post_keyword','[{"system":false,"id":"znlhv1g7","name":"post_id","type":"relation","required":false,"unique":false,"options":{"maxSelect":1,"collectionId":"d6mbq3heomws7j8","cascadeDelete":true}},{"system":false,"id":"ad808gc8","name":"keyword_id","type":"relation","required":false,"unique":false,"options":{"maxSelect":1,"collectionId":"ue9kdud987b9z5z","cascadeDelete":true}}]','','','','','','{}','2022-11-29 10:22:46.480Z','2022-12-01 13:44:31.318Z');
								
				`)

		res, errMigrate := migrateQuery.Execute()

		fmt.Println("migrate", res, errMigrate)
		return errMigrate
	}, func(db dbx.Builder) error {
		// add down queries...

		return nil
	})
}
