package example

import (
	"net/url"
	"path"
	"time"

	"github.com/morphy76/g-fe-server/pkg/example"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"context"
)

const MONGO_COLLECTION = "examples"

type MongoRepository struct {
	Url      string
	Username string
	Password string

	connected  bool
	ctx        context.Context
	client     *mongo.Client
	collection *mongo.Collection
}

func (r *MongoRepository) FindAll() ([]example.Example, error) {

	if !r.connected {
		return nil, example.ErrNotConnected
	}

	cur, err := r.collection.Find(r.ctx, bson.D{})
	if err != nil {
		return nil, err
	}

	defer cur.Close(r.ctx)

	var rv []example.Example
	err = cur.All(r.ctx, &rv)
	if err != nil {
		return nil, err
	}

	return rv, nil
}

func (r *MongoRepository) FindById(id string) (example.Example, error) {

	rv := example.Example{}

	if !r.connected {
		return rv, example.ErrNotConnected
	}

	singleResult := r.collection.FindOne(r.ctx, bson.D{{Key: "name", Value: id}})

	err := singleResult.Decode(&rv)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return rv, example.ErrNotFound
		}
		return rv, err
	}

	return rv, nil
}

func (r *MongoRepository) Save(e example.Example) error {

	if !r.connected {
		return example.ErrNotConnected
	}

	_, err := r.collection.InsertOne(r.ctx, e)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return example.ErrAlreadyExists
		}
		return err
	}

	return nil
}

func (r *MongoRepository) Update(e example.Example) error {

	if !r.connected {
		return example.ErrNotConnected
	}

	updateResult, err := r.collection.ReplaceOne(r.ctx, bson.D{{Key: "name", Value: e.Name}}, e)
	if err != nil {
		return err
	}
	if updateResult.MatchedCount == 0 {
		return example.ErrNotFound
	}

	return nil
}

func (r *MongoRepository) Delete(id string) error {

	if !r.connected {
		return example.ErrNotConnected
	}

	deleteResult, err := r.collection.DeleteOne(r.ctx, bson.D{{Key: "name", Value: id}})
	if err != nil {
		return err
	}
	if deleteResult.DeletedCount == 0 {
		return example.ErrNotFound
	}

	return nil
}

func (r *MongoRepository) Connect() error {

	r.ctx = context.Background()

	useUrl, err := url.Parse(r.Url)
	if err != nil {
		return err
	}

	var clientOpts *options.ClientOptions
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	if useUrl.User == nil {
		useCredentials := url.UserPassword(r.Username, r.Password)
		useUrl.User = useCredentials
	}

	clientOpts = options.Client().
		ApplyURI(useUrl.String()).
		SetServerAPIOptions(serverAPI)

	r.client, err = mongo.Connect(r.ctx, clientOpts)
	r.connected = err == nil

	r.collection = r.client.Database(path.Base(useUrl.Path)).Collection(MONGO_COLLECTION)

	return err
}

func (r *MongoRepository) Disconnect() error {
	var err error
	if r.connected {
		err = r.client.Disconnect(r.ctx)
	} else {
		err = example.ErrNotConnected
	}
	r.connected = false
	return err
}

func (r *MongoRepository) IsConnected() bool {
	return r.connected
}

func (r *MongoRepository) Ping() bool {
	if !r.connected {
		return false
	}

	ctx, cancel := context.WithTimeout(r.ctx, 5*time.Second)
	defer cancel()

	err := r.client.Ping(ctx, nil)
	return err == nil
}