package example

import (
	"net/url"
	"path"

	"github.com/morphy76/g-fe-server/internal/options"
	"github.com/morphy76/g-fe-server/pkg/example"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"context"
)

const MONGO_COLLECTION = "examples"

type MongoRepository struct {
	DbOptions  *options.DbOptions
	Client     *mongo.Client
	UseContext context.Context
	collection *mongo.Collection
}

func (r *MongoRepository) FindAll() ([]example.Example, error) {

	r.lazyBindCollection()

	cur, err := r.collection.Find(r.UseContext, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(r.UseContext)

	var rv []example.Example
	err = cur.All(r.UseContext, &rv)
	if err != nil {
		return nil, err
	}

	return rv, nil
}

func (r *MongoRepository) FindById(id string) (example.Example, error) {

	r.lazyBindCollection()

	rv := example.Example{}

	singleResult := r.collection.FindOne(r.UseContext, bson.D{{Key: "name", Value: id}})

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

	r.lazyBindCollection()

	_, err := r.collection.InsertOne(r.UseContext, e)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return example.ErrAlreadyExists
		}
		return err
	}

	return nil
}

func (r *MongoRepository) Update(e example.Example) error {

	r.lazyBindCollection()

	updateResult, err := r.collection.ReplaceOne(r.UseContext, bson.D{{Key: "name", Value: e.Name}}, e)
	if err != nil {
		return err
	}
	if updateResult.MatchedCount == 0 {
		return example.ErrNotFound
	}

	return nil
}

func (r *MongoRepository) Delete(id string) error {

	r.lazyBindCollection()

	deleteResult, err := r.collection.DeleteOne(r.UseContext, bson.D{{Key: "name", Value: id}})
	if err != nil {
		return err
	}
	if deleteResult.DeletedCount == 0 {
		return example.ErrNotFound
	}

	return nil
}

func (r *MongoRepository) lazyBindCollection() {
	if r.collection == nil {

		useUrl, _ := url.Parse(r.DbOptions.Url)

		if useUrl.User == nil {
			useCredentials := url.UserPassword(r.DbOptions.User, r.DbOptions.Password)
			useUrl.User = useCredentials
		}

		r.collection = r.Client.Database(path.Base(useUrl.Path)).Collection(MONGO_COLLECTION)
	}
}
