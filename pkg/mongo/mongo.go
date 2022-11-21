package mongo

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"strconv"
	"time"
)

// GetInterfaceToString convert interface{} to string type
func GetInterfaceToString(value interface{}) string {
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}

// MongoInfo mongo connect information
type MongoInfo struct {
	DB              string `json:"db,omitempty" form:"db,omitempty"`
	MongoConnectUrl string `json:"mongo_connect_url,omitempty" form:"mongo_connect_url,omitempty"`

	client *mongo.Client
}

// MongoPoolConnect init connection pool for mongo
func (m *MongoInfo) MongoPoolConnect(max uint64) *mongo.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	mongoOptions := options.Client().ApplyURI(m.MongoConnectUrl)
	mongoOptions.SetMaxPoolSize(max)

	var err error
	m.client, err = mongo.Connect(ctx, mongoOptions)
	if err != nil {
		return nil
	}
	return m.client
}

// MongoPoolDisconnect defer connection pool for mongo
func (m *MongoInfo) MongoPoolDisconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	defer func() {
		if err := m.client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

// MongoInsertMany mongo command db.col.insert_many(documents)
func (m *MongoInfo) MongoInsertMany(col string, documents []interface{}, opts ...*options.InsertManyOptions) []interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	collection := m.client.Database(m.DB).Collection(col)
	result, err := collection.InsertMany(ctx, documents, opts...)
	if err != nil {
		return nil
	}
	return result.InsertedIDs
}

// MongoInsertOne mongo command db.col.insert_one(document)
func (m *MongoInfo) MongoInsertOne(col string, document interface{}, opts ...*options.InsertOneOptions) interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	collection := m.client.Database(m.DB).Collection(col)
	result, err := collection.InsertOne(ctx, document, opts...)
	if err != nil {
		return err
	}
	return result
}

// MongoFind mongo command db.col.find(filter)
func (m *MongoInfo) MongoFind(col string, filter interface{}, opts ...*options.FindOptions) (Results []*bson.D, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	collection := m.client.Database(m.DB).Collection(col)
	cur, err := collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	Results, err = m.mongoCursor(ctx, cur)
	return
}

// MongoAggregate mongo command db.col.aggregate()
func (m *MongoInfo) MongoAggregate(col string, filter interface{}, opts ...*options.AggregateOptions) (Results []*bson.D, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	collection := m.client.Database(m.DB).Collection(col)
	cur, err := collection.Aggregate(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	Results, err = m.mongoCursor(ctx, cur)
	return
}

// mongoCursor mongo cursor read and return data
func (m *MongoInfo) mongoCursor(ctx context.Context, cur *mongo.Cursor) (Results []*bson.D, err error) {
	for cur.Next(ctx) {
		var result bson.D
		err := cur.Decode(&result)
		if err != nil {
			return nil, err
		}
		Results = append(Results, &result)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return
}

// MongoCount mongo command db.col.count(filter)
func (m *MongoInfo) MongoCount(col string, filter interface{}, opts ...*options.CountOptions) (count int64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	collection := m.client.Database(m.DB).Collection(col)
	count, err = collection.CountDocuments(ctx, filter, opts...)
	if err != nil {
		return 0, err
	}
	return
}

func (m *MongoInfo) MongoUpdateMany(col string, filter interface{}, update interface{},
	opts ...*options.UpdateOptions) (updateResult *mongo.UpdateResult, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	collection := m.client.Database(m.DB).Collection(col)
	updateResult, err = collection.UpdateMany(ctx, filter, update, opts...)
	return
}

func (m *MongoInfo) MongoUpdateById(col string, id interface{}, update interface{},
	opts ...*options.UpdateOptions) (updateResult *mongo.UpdateResult, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	collection := m.client.Database(m.DB).Collection(col)
	updateResult, err = collection.UpdateByID(ctx, id, update, opts...)
	return
}
