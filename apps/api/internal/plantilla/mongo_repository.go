package plantilla

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

var ErrNotFound = errors.New("plantilla no encontrada")
var ErrConflict = errors.New("versión esperada no coincide")

type MongoRepository struct {
	client              *mongo.Client
	templates, versions *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, db string) *MongoRepository {
	d := client.Database(db)
	return &MongoRepository{client, d.Collection("plantillas"), d.Collection("plantilla_versiones")}
}
func (r *MongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.templates.Indexes().CreateMany(ctx, []mongo.IndexModel{{Keys: bson.D{{Key: "propietario_id", Value: 1}, {Key: "deleted_at", Value: 1}, {Key: "created_at", Value: -1}, {Key: "_id", Value: -1}}}})
	if err != nil {
		return err
	}
	_, err = r.versions.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: bson.D{{Key: "plantilla_id", Value: 1}, {Key: "numero", Value: 1}}, Options: options.Index().SetUnique(true)})
	return err
}
func (r *MongoRepository) Create(ctx context.Context, owner string, in CreateRequest) (Response, error) {
	now := time.Now().UTC()
	id := bson.NewObjectID()
	active := true
	if in.Activo != nil {
		active = *in.Activo
	}
	d := document{ID: id, Owner: owner, Current: 1, Created: now, Updated: now}
	v := version{TemplateID: id, Number: 1, Name: in.Nombre, Description: in.Descripcion, Content: in.Contenido, Active: active, Author: owner, Created: now}
	s, err := r.client.StartSession()
	if err != nil {
		return Response{}, err
	}
	defer s.EndSession(ctx)
	_, err = s.WithTransaction(ctx, func(sc context.Context) (any, error) {
		if _, e := r.templates.InsertOne(sc, d); e != nil {
			return nil, e
		}
		if _, e := r.versions.InsertOne(sc, v); e != nil {
			return nil, e
		}
		return nil, nil
	})
	if err != nil {
		return Response{}, err
	}
	return response(d, v), nil
}
func (r *MongoRepository) filter(owner string, all bool, id string) (bson.D, error) {
	oid, e := bson.ObjectIDFromHex(id)
	if e != nil {
		return nil, fmt.Errorf("id inválido")
	}
	f := bson.D{{Key: "_id", Value: oid}, {Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}}}
	if !all {
		f = append(f, bson.E{Key: "propietario_id", Value: owner})
	}
	return f, nil
}
func (r *MongoRepository) current(ctx context.Context, owner string, all bool, id string) (document, version, error) {
	f, e := r.filter(owner, all, id)
	if e != nil {
		return document{}, version{}, e
	}
	var d document
	if e = r.templates.FindOne(ctx, f).Decode(&d); e != nil {
		if errors.Is(e, mongo.ErrNoDocuments) {
			return d, version{}, ErrNotFound
		}
		return d, version{}, e
	}
	var v version
	e = r.versions.FindOne(ctx, bson.D{{Key: "plantilla_id", Value: d.ID}, {Key: "numero", Value: d.Current}}).Decode(&v)
	return d, v, e
}
func (r *MongoRepository) Get(ctx context.Context, o string, a bool, id string) (Response, error) {
	d, v, e := r.current(ctx, o, a, id)
	if e != nil {
		return Response{}, e
	}
	return response(d, v), nil
}
func (r *MongoRepository) List(ctx context.Context, o string, all bool, page, limit int) ([]Response, int64, error) {
	f := bson.D{{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}}}
	if !all {
		f = append(f, bson.E{Key: "propietario_id", Value: o})
	}
	total, e := r.templates.CountDocuments(ctx, f)
	if e != nil {
		return nil, 0, e
	}
	cur, e := r.templates.Find(ctx, f, findOptions(page, limit))
	if e != nil {
		return nil, 0, e
	}
	defer cur.Close(ctx)
	out := []Response{}
	for cur.Next(ctx) {
		var d document
		if e = cur.Decode(&d); e != nil {
			return nil, 0, e
		}
		var v version
		e = r.versions.FindOne(ctx, bson.D{{Key: "plantilla_id", Value: d.ID}, {Key: "numero", Value: d.Current}}).Decode(&v)
		if e != nil {
			return nil, 0, e
		}
		out = append(out, response(d, v))
	}
	return out, total, cur.Err()
}
func findOptions(page, limit int) *options.FindOptionsBuilder {
	return options.Find().SetSkip(int64((page - 1) * limit)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: -1}})
}
func (r *MongoRepository) Update(ctx context.Context, o string, all bool, id string, in UpdateRequest) (Response, error) {
	d, _, e := r.current(ctx, o, all, id)
	if e != nil {
		return Response{}, e
	}
	if d.Current != in.VersionEsperada {
		return Response{}, ErrConflict
	}
	now := time.Now().UTC()
	next := d.Current + 1
	v := version{TemplateID: d.ID, Number: next, Name: in.Nombre, Description: in.Descripcion, Content: in.Contenido, Active: *in.Activo, Author: o, Created: now}
	s, e := r.client.StartSession()
	if e != nil {
		return Response{}, e
	}
	defer s.EndSession(ctx)
	_, e = s.WithTransaction(ctx, func(sc context.Context) (any, error) {
		res, e := r.templates.UpdateOne(sc, bson.D{{Key: "_id", Value: d.ID}, {Key: "version_actual", Value: in.VersionEsperada}, {Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}}}, bson.D{{Key: "$set", Value: bson.D{{Key: "version_actual", Value: next}, {Key: "updated_at", Value: now}}}})
		if e != nil {
			return nil, e
		}
		if res.MatchedCount == 0 {
			return nil, ErrConflict
		}
		if _, e = r.versions.InsertOne(sc, v); e != nil {
			return nil, e
		}
		return nil, nil
	})
	if e != nil {
		return Response{}, e
	}
	d.Current = next
	d.Updated = now
	return response(d, v), nil
}
func (r *MongoRepository) Delete(ctx context.Context, o string, a bool, id string) error {
	f, e := r.filter(o, a, id)
	if e != nil {
		return e
	}
	now := time.Now().UTC()
	res, e := r.templates.UpdateOne(ctx, f, bson.D{{Key: "$set", Value: bson.D{{Key: "deleted_at", Value: now}}}})
	if e != nil {
		return e
	}
	if res.MatchedCount == 0 {
		return ErrNotFound
	}
	return nil
}
func (r *MongoRepository) Versions(ctx context.Context, o string, a bool, id string, page, limit int) ([]VersionResponse, int64, error) {
	d, _, e := r.current(ctx, o, a, id)
	if e != nil {
		return nil, 0, e
	}
	f := bson.D{{Key: "plantilla_id", Value: d.ID}}
	total, e := r.versions.CountDocuments(ctx, f)
	if e != nil {
		return nil, 0, e
	}
	cur, e := r.versions.Find(ctx, f, findOptions(page, limit))
	if e != nil {
		return nil, 0, e
	}
	defer cur.Close(ctx)
	out := []VersionResponse{}
	for cur.Next(ctx) {
		var v version
		if e = cur.Decode(&v); e != nil {
			return nil, 0, e
		}
		out = append(out, versionResponse(v))
	}
	return out, total, cur.Err()
}
func (r *MongoRepository) Version(ctx context.Context, o string, a bool, id string, n int) (VersionResponse, error) {
	d, _, e := r.current(ctx, o, a, id)
	if e != nil {
		return VersionResponse{}, e
	}
	var v version
	e = r.versions.FindOne(ctx, bson.D{{Key: "plantilla_id", Value: d.ID}, {Key: "numero", Value: n}}).Decode(&v)
	if errors.Is(e, mongo.ErrNoDocuments) {
		return VersionResponse{}, ErrNotFound
	}
	return versionResponse(v), e
}
func response(d document, v version) Response {
	return Response{ID: d.ID.Hex(), PropietarioID: d.Owner, VersionActual: d.Current, VersionResponse: versionResponse(v), CreatedAt: d.Created, UpdatedAt: d.Updated}
}
func versionResponse(v version) VersionResponse {
	return VersionResponse{Numero: v.Number, Nombre: v.Name, Descripcion: v.Description, Contenido: v.Content, Activo: v.Active, AutorID: v.Author, CreatedAt: v.Created}
}
