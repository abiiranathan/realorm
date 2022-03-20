package realorm_test

import (
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/abiiranathan/realorm/database"
	"github.com/abiiranathan/realorm/realorm"
)

type Post struct {
	ID      uint   `gorm:"primary_key;not null;AUTO_INCREMENT"`
	Title   string `gorm:"type:varchar(100);not null"`
	Content string `gorm:"type:varchar(1000);not null"`
}

// Setup
// setup the database

func UniqueID() uint {
	// Generate a unique ID
	rand.Seed(time.Now().UnixNano())
	return uint(rand.Intn(math.MaxInt32))
}

func clear_table(t *testing.T) {
	orm, err := create_orm()
	if err != nil {
		t.Errorf("error creating database: %v\n", err)
	}

	orm.GetDB().Exec("DELETE FROM posts;")
}

func create_orm() (realorm.ORM, error) {
	orm := realorm.New(database.SQLITE3_MEMORY_DB, database.SQLITE3)
	err := orm.Migrate(&Post{})
	return orm, err
}

func create_post(t *testing.T) (*Post, realorm.ORM) {
	orm, err := create_orm()
	if err != nil {
		t.Errorf("error creating database: %v\n", err)
	}

	post := &Post{
		ID:      UniqueID(),
		Title:   "Hello World",
		Content: "This is a test post",
	}

	err = orm.Create(post)
	if err != nil {
		t.Errorf("error creating post: %v\n", err)
	}

	// test create with wrong type
	err = orm.Create(1)
	if err == nil {
		t.Errorf("expected error due to wrong type, got nil")
	}

	return post, orm
}

func Test_realorm(t *testing.T) {
	t.Parallel()

	Test_realorm_FindAllPaginated(t)
	Test_realorm_FindAll(t)
	Test_realorm_Find(t)
	Test_realorm_Create(t)
	Test_realorm_Update(t)
	Test_realorm_Delete(t)
}

func Test_realorm_Find(t *testing.T) {
	defer clear_table(t)

	post, orm := create_post(t)
	var post2 Post
	err := orm.Find(&post2, &realorm.WhereClause{
		Query: "id = ?",
		Args:  []interface{}{post.ID},
	})

	if err != nil {
		t.Errorf("error finding post: %v\n", err)
	}

	if post2.ID != post.ID {
		t.Errorf("post id mismatch: %d != %d\n", post2.ID, post.ID)
	}

	// test find with no where clause
	var post3 Post
	err = orm.Find(&post3, nil)
	if err == nil {
		t.Errorf("this should fail because the where clause is required: %v\n", err)
	}

	// error should be a realorm.ErrNoWhereClause
	if err != realorm.ErrNoWhereClause {
		t.Errorf("error should be a realorm.ErrNoWhereClause: %v\n", err)
	}

}

func Test_realorm_FindAllPaginated(t *testing.T) {
	defer clear_table(t)

	orm, err := create_orm()
	if err != nil {
		t.Errorf("error creating database: %v\n", err)
	}

	var posts []Post
	var paginatedResult *realorm.PaginatedResult
	paginatedResult, _ = orm.FindAllPaginated(&posts, 1, 10, nil)

	if paginatedResult.TotalPages != 0 {
		t.Errorf("expected 0, got %d", paginatedResult.TotalPages)
	}

	// test paginate with where clause
	post, _ := create_post(t)
	paginatedResult, _ = orm.FindAllPaginated(&posts, 1, 10, &realorm.WhereClause{
		Query: "id = ?",
		Args:  []interface{}{post.ID},
	})

	if paginatedResult.TotalPages != 1 {
		t.Errorf("expected 1, got %d", paginatedResult.TotalPages)
	}

	// interface conversion, comma-ok
	if newPosts, ok := paginatedResult.Results.(*[]Post); !ok {
		t.Errorf("expected []Post, got %T", paginatedResult.Results)
	} else {
		if len(*newPosts) != 1 {
			t.Errorf("expected 1, got %d", len(*newPosts))
		}

		if (*newPosts)[0].ID != post.ID {
			t.Errorf("expected %d, got %d", post.ID, (*newPosts)[0].ID)
		}

	}

	// Test with page > 1
	paginatedResult, _ = orm.FindAllPaginated(&posts, 2, 10, &realorm.WhereClause{
		Query: "id = ?",
		Args:  []interface{}{post.ID},
	})

	if paginatedResult.TotalPages != 1 {
		t.Errorf("expected 1, got %d", paginatedResult.TotalPages)
	}

	// Test invalid column name to return error
	_, err = orm.FindAllPaginated(&posts, 1, 10, &realorm.WhereClause{
		Query: "invalid_column = ?",
		Args:  []interface{}{post.ID},
	})

	if err == nil {
		t.Errorf("expected error due to invalid column, got nil")
	}
}

func Test_realorm_FindAll(t *testing.T) {
	defer clear_table(t)

	orm, err := create_orm()
	if err != nil {
		t.Errorf("error creating database: %v\n", err)
	}

	var posts []Post
	err = orm.FindAll(&posts, nil)

	if err != nil {
		t.Errorf("error finding all posts: %v\n", err)
	}

	// test paginate with where clause
	post, _ := create_post(t)
	err = orm.FindAll(&posts, &realorm.WhereClause{
		Query: "id = ?",
		Args:  []interface{}{post.ID},
	})

	if err != nil {
		t.Errorf("error finding all posts: %v\n", err)
	}

	if len(posts) != 1 {
		t.Errorf("expected 1, got %d", len(posts))
	}

	if posts[0].ID != post.ID {
		t.Errorf("expected %d, got %d", post.ID, posts[0].ID)
	}

	err = orm.FindAll(&posts, &realorm.WhereClause{
		Query: "invalid_column = ?",
		Args:  []interface{}{post.ID},
	})

	if err == nil {
		t.Errorf("expected error due to invalid column, got nil")
	}

}

func Test_realorm_Create(t *testing.T) {
	defer clear_table(t)

	_, _ = create_post(t)
}

func Test_realorm_Update(t *testing.T) {
	defer clear_table(t)

	post, orm := create_post(t)
	post.Title = "Hello World 2"

	data, err := orm.Update(*post, post.ID, &realorm.WhereClause{
		Query: "id = ?",
		Args:  []interface{}{post.ID},
	})

	if err != nil {
		t.Errorf("error updating post: %v\n", err)
	}

	updatedPost := data.(*Post)

	if updatedPost.Title != post.Title {
		t.Errorf("post title mismatch: %s != %s\n", updatedPost.Title, post.Title)
	}

	// test update with no where clause
	post.Title = "Hello World 3"
	data, err = orm.Update(*post, post.ID, nil)

	if err == nil {
		// this should fail because the where clause is required
		t.Errorf("this should fail because the where clause is required: %v\n", err)
	}

	// error should be a realorm.ErrNoWhereClause
	if err != realorm.ErrNoWhereClause {
		t.Errorf("error should be a realorm.ErrNoWhereClause: %v\n", err)
	}

	// test update with a column that doesn't exist
	// 404
	post.Title = "Hello World 4"
	_, err = orm.Update(*post, post.ID+100, &realorm.WhereClause{
		Query: "id = ?",
		Args:  []interface{}{post.ID + 100},
	})

	if err == nil {
		// this should fail because the where clause is required
		t.Errorf("this should fail because the row does not exist: %v\n", err)
	}

	// Invalid column name
	post.Title = "Hello World 5"
	_, err = orm.Update(*post, post.ID, &realorm.WhereClause{
		Query: "invalid_column = ?",
		Args:  []interface{}{post.ID},
	})

	if err == nil {
		// this should fail because the where clause is required
		t.Errorf("this should fail because of invalid column name: %v\n", err)
	}

}

func Test_realorm_Delete(t *testing.T) {
	defer clear_table(t)

	post, orm := create_post(t)

	err := orm.Delete(&Post{}, &realorm.WhereClause{
		Query: "id = ?",
		Args:  []interface{}{post.ID},
	})

	if err != nil {
		t.Errorf("error deleting post: %v\n", err)
	}

	// test delete with no where clause
	err = orm.Delete(&Post{}, nil)

	if err == nil {
		// this should fail because the where clause is required
		t.Errorf("this should fail because the where clause is required: %v\n", err)
	}
}
