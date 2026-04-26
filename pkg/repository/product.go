package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Product struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name      string             `bson:"name" json:"name"`
	Price     float64            `bson:"price" json:"price"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

type ProductRepository interface {
	FindAll(ctx context.Context) ([]Product, error)
	FindByID(ctx context.Context, id primitive.ObjectID) (*Product, error)
	Insert(ctx context.Context, p *Product) (*Product, error)
	Update(ctx context.Context, id primitive.ObjectID, p *Product) error
	Delete(ctx context.Context, id primitive.ObjectID) error
}

type MongoProductRepository struct {
	collection *mongo.Collection
}

func NewMongoProductRepository(db *mongo.Database, collectionName string) *MongoProductRepository {
	return &MongoProductRepository{
		collection: db.Collection(collectionName),
	}
}

func (r *MongoProductRepository) FindAll(ctx context.Context) ([]Product, error) {
	cursor, err := r.collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []Product
	if err = cursor.All(ctx, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *MongoProductRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*Product, error) {
	var product Product
	filter := bson.M{"_id": id}
	err := r.collection.FindOne(ctx, filter).Decode(&product)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *MongoProductRepository) Insert(ctx context.Context, p *Product) (*Product, error) {
	p.ID = primitive.NewObjectID()
	p.CreatedAt = time.Now().UTC()

	_, err := r.collection.InsertOne(ctx, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *MongoProductRepository) Update(ctx context.Context, id primitive.ObjectID, p *Product) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{
		"name":  p.Name,
		"price": p.Price,
	}}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *MongoProductRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}
