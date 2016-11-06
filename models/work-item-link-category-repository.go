package models

import (
	"log"

	"golang.org/x/net/context"

	"github.com/almighty/almighty-core/app"
	"github.com/jinzhu/gorm"
	satoriuuid "github.com/satori/go.uuid"
)

// NewWorkItemLinkCategoryRepository creates a work item link category repository based on gorm
func NewWorkItemLinkCategoryRepository(db *gorm.DB) *GormWorkItemLinkCategoryRepository {
	return &GormWorkItemLinkCategoryRepository{db}
}

// GormWorkItemLinkCategoryRepository implements WorkItemLinkCategoryRepository using gorm
type GormWorkItemLinkCategoryRepository struct {
	db *gorm.DB
}

// Create creates a new work item link category in the repository.
// Returns BadParameterError, ConversionError or InternalError
func (r *GormWorkItemLinkCategoryRepository) Create(ctx context.Context, name *string, description *string) (*app.WorkItemLinkCategory, error) {

	if name == nil || *name == "" {
		return nil, BadParameterError{parameter: "name", value: name}
	}

	created := WorkItemLinkCategory{
		// Omit "lifecycle" and "ID" fields as they will be filled by the DB
		Name:        *name,
		Description: description,
	}

	db := r.db.Create(&created)
	if db.Error != nil {
		return nil, InternalError{simpleError{db.Error.Error()}}
	}

	// Convert the created link category entry into a JSONAPI response
	result := convertLinkCategoryFromModel(&created)

	return &result, nil
}

// Load returns the work item link category for the given ID.
// Returns NotFoundError, ConversionError or InternalError
func (r *GormWorkItemLinkCategoryRepository) Load(ctx context.Context, ID string) (*app.WorkItemLinkCategory, error) {
	id, err := satoriuuid.FromString(ID)
	if err != nil {
		// treat as not found: clients don't know it must be a UUID
		return nil, NotFoundError{entity: "work item link category", ID: ID}
	}
	log.Printf("loading work item link category %s", id.String())
	res := WorkItemLinkCategory{}
	db := r.db.First(&res, id)
	if db.RecordNotFound() {
		log.Printf("not found, res=%v", res)
		return nil, NotFoundError{"work item link category", id.String()}
	}

	// Convert the created link category entry into a JSONAPI response
	result := convertLinkCategoryFromModel(&res)
	return &result, nil
}

// List returns all work item link categories
// TODO: Handle pagination
func (r *GormWorkItemLinkCategoryRepository) List(ctx context.Context) (*app.WorkItemLinkCategoryArray, error) {

	// We don't have any where clause or paging at the moment.
	var where string
	var parameters []interface{}
	var start *int
	var limit *int

	var rows []WorkItemLinkCategory
	db := r.db.Where(where, parameters...)
	if start != nil {
		db = db.Offset(*start)
	}
	if limit != nil {
		db = db.Limit(*limit)
	}
	db = db.Find(&rows)
	if db.Error != nil {
		return nil, db.Error
	}
	res := app.WorkItemLinkCategoryArray{}
	res.Data = make([]*app.WorkItemLinkCategory, len(rows))

	for index, value := range rows {
		cat := convertLinkCategoryFromModel(&value)
		res.Data[index] = &cat
	}

	// TODO: When adding pagination, this must not be len(rows) but
	// the overall total number of elements from all pages.
	res.Meta = &app.WorkItemLinkCategoryArrayMeta{
		TotalCount: len(rows),
	}

	return &res, nil
}

// Delete deletes the work item link category with the given id
// returns NotFoundError or InternalError
func (r *GormWorkItemLinkCategoryRepository) Delete(ctx context.Context, ID string) error {
	id, err := satoriuuid.FromString(ID)
	if err != nil {
		// treat as not found: clients don't know it must be a UUID
		return NotFoundError{entity: "work item link category", ID: ID}
	}

	var cat = WorkItemLinkCategory{
		ID: id,
	}

	log.Printf("work item link category to delete %v\n", cat)

	db := r.db.Delete(&cat)
	if db.Error != nil {
		return InternalError{simpleError{db.Error.Error()}}
	}

	if db.RowsAffected == 0 {
		return NotFoundError{entity: "work item link category", ID: id.String()}
	}
	return nil
}

// Save updates the given work item link category in storage. Version must be the same as the one int the stored version.
// returns NotFoundError, VersionConflictError, ConversionError or InternalError
func (r *GormWorkItemLinkCategoryRepository) Save(ctx context.Context, linkCat app.WorkItemLinkCategory) (*app.WorkItemLinkCategory, error) {
	res := WorkItemLinkCategory{}
	id, err := satoriuuid.FromString(linkCat.Data.ID)
	if err != nil {
		log.Printf("Error when converting %s to UUID: %s", linkCat.Data.ID, err.Error())
		// treat as not found: clients don't know it must be a UUID
		return nil, NotFoundError{entity: "work item link category", ID: id.String()}
	}

	if linkCat.Data.Type != "workitemlinkcategories" {
		return nil, BadParameterError{parameter: "data.type", value: linkCat.Data.Type}
	}

	// If the name is not nil, it MUST NOT be empty
	if linkCat.Data.Attributes.Name != nil && *linkCat.Data.Attributes.Name == "" {
		return nil, BadParameterError{parameter: "data.attributes.name", value: *linkCat.Data.Attributes.Name}
	}

	db := r.db.First(&res, id)
	if db.RecordNotFound() {
		log.Printf("not found, res=%v", res)
		return nil, NotFoundError{entity: "work item link category", ID: id.String()}
	}
	if linkCat.Data.Attributes.Version == nil || res.Version != *linkCat.Data.Attributes.Version {
		return nil, VersionConflictError{simpleError{"version conflict"}}
	}

	description := ""
	if linkCat.Data.Attributes.Description != nil {
		description = *linkCat.Data.Attributes.Description
	}

	name := ""
	if linkCat.Data.Attributes.Name != nil {
		name = *linkCat.Data.Attributes.Name
	}

	newLinkCat := WorkItemLinkCategory{
		ID:          id,
		Name:        name,
		Description: &description,
		Version:     *linkCat.Data.Attributes.Version + 1,
	}

	db = db.Save(&newLinkCat)
	if db.Error != nil {
		log.Print(db.Error.Error())
		return nil, InternalError{simpleError{db.Error.Error()}}
	}
	log.Printf("updated work item link category to %v\n", newLinkCat)
	result := convertLinkCategoryFromModel(&newLinkCat)
	return &result, nil
}

// convertLinkCategoryFromModel converts from model to app representation
func convertLinkCategoryFromModel(t *WorkItemLinkCategory) app.WorkItemLinkCategory {
	var converted = app.WorkItemLinkCategory{
		Data: &app.WorkItemLinkCategoryData{
			Type: "workitemlinkcategories",
			ID:   t.ID.String(),
			Attributes: &app.WorkItemLinkCategoryAttributes{
				Name:        &t.Name,
				Description: t.Description,
				Version:     &t.Version,
			},
		},
	}
	return converted
}