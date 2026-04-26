package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// User MongoDB-də saxlanacaq istifadəçi sənədini təmsil edir.
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username"      json:"username"`
	Email     string             `bson:"email"         json:"email"`
	Password  string             `bson:"password"      json:"-"` // JSON-da heç vaxt göstərilməsin
	CreatedAt time.Time          `bson:"created_at"    json:"created_at"`
}

// UserRepository istifadəçi üzərində əməliyyatlar interfeysi.
type UserRepository interface {
	FindByUsername(ctx context.Context, username string) (*User, error)
	Create(ctx context.Context, user *User) (*User, error)
}

// MongoUserRepository UserRepository-i MongoDB ilə reallaşdırır.
type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(db *mongo.Database, collectionName string) *MongoUserRepository {
	return &MongoUserRepository{collection: db.Collection(collectionName)}
}

func (r *MongoUserRepository) Create(ctx context.Context, user *User) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user.Password = string(hashedPassword)
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now().UTC()

	_, err = r.collection.InsertOne(ctx, user)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *MongoUserRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	filter := bson.M{"username": username}
	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
