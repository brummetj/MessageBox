// Dataservice controller for couchbase. Handles all queries, connections, and cluster initialization

package dataservice

import (
	. "MessageBox/config"
	. "MessageBox/util/logger"
	"fmt"
	"github.com/couchbase/gocb/v2"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"log"
	"time"
)

var CB CouchBaseService

func NewCouchBaseDB(config AppConfig){
	Log.Info("Connecting to datastore")
	db := CouchBaseService{}
	err := db.Init(config)
	if err != nil {
		Log.Warnf("Failure to connect to DB! Reason: %v", err)
		Log.Info("Database reconnects can be retried or reconnected via endpoint")
	}

	CB = db
}

type CouchBaseService struct {
	BucketName string
	Bucket *gocb.Bucket
	Cluster *gocb.Cluster
	CollectionName string
	Collection gocb.Collection
	Host string
	Username string
	Password string
}

func (cb *CouchBaseService) Init(config AppConfig) error {
	cb.Host = config.CouchBaseConfig.Host
	cb.Username = config.CouchBaseConfig.Username
	cb.Password = config.CouchBaseConfig.Password
	cb.BucketName = config.CouchBaseConfig.Bucket

	err := retry(config.CouchBaseConfig.Retries, 20, func() (err error) {
		err = cb.Connect()
		err = cb.OpenBucket(cb.BucketName)
		return
	})
	Log.Info("Couchbase connection success. Cluster object initialized")
	Log.Info("Bucket connection success. Bucket object initialized")
	if err != nil {
		Log.Warnf("Unable to connect to the data store! " +
			"Please check connections to make sure CB is successfully up and running")
		return err
	}

	cb.GetDefaultCollection()
	err = cb.createIndex()
	if err != nil {
		return err
	}
	return nil
}

func (cb* CouchBaseService) createIndex() error {
	indexQuery := "SELECT * FROM system:indexes"
	result, err := cb.Query(indexQuery, false)
	if err != nil {
		Log.Warnf("Error querying indexes: %v", err)
	}
	if result != nil {
		Log.Debug("Index values already created")
		return nil
	}
	mgr := cb.Cluster.QueryIndexes()
	err = mgr.CreatePrimaryIndex(cb.BucketName, nil)
	if err != nil {
		return errors.Wrap(err, "Creating index")
	}
	err = mgr.CreateIndex(cb.BucketName, "index_username", []string{"username"}, nil)
	if err != nil {
		return errors.Wrap(err, "Creating username index")
	}
	return nil
}

func (cb *CouchBaseService) createUUID() string {
	id := uuid.New()
	return id.String()
}

func (cb *CouchBaseService) Connect() error {
	cluster, err := gocb.Connect(
		cb.Host,
		gocb.ClusterOptions{Username: cb.Username, Password: cb.Password,
		})
	if err != nil {
		return errors.Wrap(err, "Connecting to couchbase")
	}
	cb.Cluster = cluster
	return nil
}

func (cb *CouchBaseService) Query(queryStatement string, multi bool) (interface{}, error) {
	query, err := cb.Cluster.Query(queryStatement, nil)
	if err != nil {
		return nil, errors.Wrap(err, "query")
	}

	var result interface{}
	if multi {
		result = cb.getRows(*query)
	} else {
		result = cb.getRow(*query)
	}
	err = query.Err()
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	return result, nil
}

func (cb *CouchBaseService) getRows(query gocb.QueryResult) interface{} {
	var resultArr []interface{}
	var result interface{}
	for query.Next() {
		if err := query.Row(&result); err != nil {
			return err
		}
		resultArr = append(resultArr, result)
		Log.Debugf("Row: %v", result)
	}
	return resultArr
}

func (cb *CouchBaseService) getRow(query gocb.QueryResult) interface{} {
	var result interface{}
	for query.Next() {
		err := query.Row(&result)
		if err != nil {
			return err
		}
		Log.Debugf("Row: %v", result)
	}
	return result
}

func (cb *CouchBaseService) GetUser(user string) (interface{}, error) {
	query := fmt.Sprintf("SELECT b.*, meta(b).id FROM `%v` b WHERE username='%v'", cb.BucketName, user)
	result, err := cb.Query(query, false)
	if err != nil {
		return nil, errors.Wrap(err, "Get user")
	}
	return result, nil
}

func (cb *CouchBaseService) GetGroup(group string) (interface{}, error) {
	query := fmt.Sprintf("SELECT b.*, meta(b).id FROM `%v` b WHERE groupname='%v'", cb.BucketName, group)
	result, err := cb.Query(query, false)
	if err != nil {
		return nil, errors.Wrap(err, "Get group")
	}
	return result, nil
}

func (cb *CouchBaseService) GetMessage(messageId int, replies bool) (interface{}, error) {
	query := ""
	if replies {
		query = fmt.Sprintf("SELECT m.* FROM %v AS b UNNEST b.messages AS m WHERE m.id=%v", cb.BucketName,  messageId)
	} else {
		query = fmt.Sprintf("SELECT m.* FROM %v AS b UNNEST b.messages AS m WHERE m.id=%v AND m.re=0", cb.BucketName,  messageId)
	}
	result, err := cb.Query(query, replies)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	return result, err
}

func (cb *CouchBaseService) GetUsersInGroup(group string) (interface{}, error) {

	query := fmt.Sprintf("SELECT b.*, meta(b).id FROM `%v` b WHERE groupname='%v'", cb.BucketName, group)
	result, err := cb.Query(query, false)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	resultMap, _ := result.(map[string]interface{})
	resultSlice, _ := resultMap["usernames"]

	// Ehh i don't love this but it'll do for now ( show me a better way! )
	qrStr := "["
	for _, elm := range resultSlice.([]interface{}) {
		qrStr += fmt.Sprintf("'%v',",elm)
	}
	if last := len(qrStr) - 1; last >= 0 && qrStr[last] == ',' {
		qrStr = qrStr[:last]
	}
	qrStr += "]"

	query = fmt.Sprintf(`SELECT b.*, meta(b).id FROM %v AS b  WHERE  b.username IN %v`, cb.BucketName, qrStr)
	result, err = cb.Query(query, true)
	if err != nil {
		return nil, errors.Wrap(err, "query error")
	}
	return result, nil
}

func (cb *CouchBaseService) OpenBucket(bucketName string) error {
	bucket := cb.Cluster.Bucket(bucketName)
	err := bucket.WaitUntilReady(5 *time.Second, nil)
	if err != nil {
		return errors.Wrap(err, "Connecting to bucket")
	}
	cb.Bucket = bucket
	return nil
}

func (cb  *CouchBaseService) GetDefaultCollection() {
	collection := cb.Bucket.DefaultCollection()
	cb.Collection = *collection
}

func (cb *CouchBaseService) Insert(obj interface{}) (interface{}, error) {
	result, err := cb.Collection.Insert(cb.createUUID(), &obj, nil)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (cb *CouchBaseService) Update(id string, obj interface{}) (interface{}, error) {
	result, err := cb.Collection.Upsert(id, &obj, nil)
	if err != nil {
		return nil, err
	}
	return result, err
}

func retry(attempts int, sleep time.Duration, f func() error) (err error) {
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return
		}

		if i >= (attempts - 1) {
			break
		}
		time.Sleep(sleep)
		log.Println("retrying after error:", err)
	}
	return fmt.Errorf("after %d attempts, last error: %s", attempts, err)
}