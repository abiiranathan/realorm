package realorm

import (
	"errors"
	"reflect"

	"github.com/abiiranathan/realorm/database"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrNoWhereClause = errors.New("where clause is required")
)

func GetType(model any) any {
	return reflect.New(reflect.TypeOf(model)).Interface()
}

type PaginatedResult struct {
	// Pointer to a slice of models
	Results any `json:"results"`
	// Total number of results
	Count int64 `json:"count"`
	// Total number of pages
	TotalPages int `json:"total_pages"`
	// HasNext indicates if there are more pages
	HasNext bool `json:"has_next"`
	// HasPrev indicates if there are previous pages
	HasPrev bool `json:"has_prev"`
	// Current page
	Page int `json:"page"`
}

type WhereClause struct {
	Query string
	Args  []interface{}
}

type Finder interface {
	/*
		Find a single entity filtered by where clause
	*/
	Find(model any, where *WhereClause) error
}

type FindAllPaginater interface {
	/*
		Finds all entities for a model and returns a paginated result
		Items are filtered by the where clause if where is not nil
	*/
	FindAllPaginated(models any, page int, pageSize int, where *WhereClause) (*PaginatedResult, error)
}

type FindAller interface {
	/*
		Finds all entities for model filtered on where clause(if where is not nil)
	*/
	FindAll(model any, where *WhereClause) error
}

type Creatable interface {
	// Inserts model struct pointer into the database
	Create(model any) error
}

type Updatable interface {
	// Updates model struct in the database on primary key id
	Update(updates any, id uint, where *WhereClause) (any, error)
}

type Deletable interface {
	// Delete model where id matches the specified id
	// Model must be a pointer a valid struct
	Delete(model any, where *WhereClause) error
}

// Abstract interface for the ORM.
// Implements all interfaces for creating, reading, updating and deleting the
// database.
type BaseModel interface {
	Finder
	FindAller
	FindAllPaginater
	Creatable
	Updatable
	Deletable
}

type ORM interface {
	BaseModel

	GetDB() *gorm.DB
	Migrate(models ...interface{}) error
}

type orm struct {
	DB *gorm.DB
}

// Connect to the database with the specified dialect and connection string(dsn)
// and returns an ORM interface for the database.
// It panics if the database cannot be connected to.
// The dsn is the connection string for the database or for postgres a pointer to the database.config
// object
func New(dsn any, dialect database.DialectString) ORM {
	db, err := database.Connect(dsn, dialect)
	if err != nil {
		panic(err)
	}

	return &orm{db}
}

func (o *orm) Find(model any, where *WhereClause) error {
	if where == nil {
		return ErrNoWhereClause
	}

	return o.DB.Preload(clause.Associations).Where(where.Query, where.Args...).First(model).Error

}

func (o *orm) FindAll(models any, where *WhereClause) error {
	if where != nil {
		return o.DB.Preload(clause.Associations).Where(where.Query, where.Args...).Find(models).Error
	} else {
		return o.DB.Preload(clause.Associations).Find(models).Error
	}
}

func (o *orm) FindAllPaginated(models any, page int, pageSize int, where *WhereClause) (*PaginatedResult, error) {
	var count int64
	var err error

	if where != nil {
		err = o.DB.Model(models).Where(where.Query, where.Args...).Count(&count).Error
	} else {
		err = o.DB.Model(models).Count(&count).Error
	}

	if err != nil {
		return nil, err
	}

	totalPages := int(count) / pageSize

	if int(count)%pageSize > 0 {
		totalPages++
	}

	var offset int

	if page == 1 {
		offset = 0
	} else {
		offset = (page - 1) * pageSize
	}

	if where != nil {
		err = o.DB.Preload(clause.Associations).Model(models).Where(where.Query, where.Args...).Offset(offset).Limit(pageSize).Find(models).Error
	} else {
		err = o.DB.Preload(clause.Associations).Model(models).Offset(offset).Limit(pageSize).Find(models).Error
	}

	return &PaginatedResult{
		Results:    models,
		Count:      count,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
		Page:       page,
	}, err

}

func (o *orm) Create(model any) error {
	err := o.DB.Model(model).Create(model).Error
	if err != nil {
		return err
	}

	// refetch the model
	return o.DB.Preload(clause.Associations).First(&model, model).Error

}

func (o *orm) Update(updates any, id uint, where *WhereClause) (any, error) {
	entity := GetType(updates)

	if where == nil {
		return nil, ErrNoWhereClause
	}

	err := o.DB.First(&entity, id).Error

	if err != nil {
		return nil, err
	}

	// Update the model
	err = o.DB.Model(&entity).Where(where.Query, where.Args...).Updates(updates).Error

	if err != nil {
		return nil, err
	}

	// refetch the model
	err = o.DB.Preload(clause.Associations).Where(where.Query, where.Args...).First(&entity).Error
	return entity, err

}

func (o *orm) Delete(model any, where *WhereClause) error {
	if where == nil {
		return ErrNoWhereClause
	}

	return o.DB.Where(where.Query, where.Args).Delete(model).Error
}

func (o *orm) GetDB() *gorm.DB {
	return o.DB
}

func (o *orm) Migrate(models ...interface{}) error {
	return o.DB.AutoMigrate(models...)
}
