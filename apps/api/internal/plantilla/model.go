package plantilla

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"time"
)

type document struct {
	ID      bson.ObjectID `bson:"_id,omitempty"`
	Owner   string        `bson:"propietario_id"`
	Current int           `bson:"version_actual"`
	Created time.Time     `bson:"created_at"`
	Updated time.Time     `bson:"updated_at"`
	Deleted *time.Time    `bson:"deleted_at,omitempty"`
}
type version struct {
	ID          bson.ObjectID  `bson:"_id,omitempty"`
	TemplateID  bson.ObjectID  `bson:"plantilla_id"`
	Number      int            `bson:"numero"`
	Name        string         `bson:"nombre"`
	Description string         `bson:"descripcion"`
	Content     map[string]any `bson:"contenido"`
	Active      bool           `bson:"activo"`
	Author      string         `bson:"autor_id"`
	Created     time.Time      `bson:"created_at"`
}
