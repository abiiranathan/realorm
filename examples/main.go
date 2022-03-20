package main

import (
	"fmt"
	"log"

	"flag"

	"github.com/abiiranathan/realorm/database"
	"github.com/abiiranathan/realorm/realorm"
)

type Post struct {
	ID      uint   `gorm:"primary_key;not null;AUTO_INCREMENT"`
	Title   string `gorm:"type:varchar(100);not null"`
	Content string `gorm:"type:varchar(1000);not null"`
}

var (
	// DSN is the database connection string
	DSN     = flag.String("dsn", "", "database connection string")
	MIGRATE = flag.Bool("migrate", false, "migrate the database schema")
)

func main() {
	flag.Parse()

	// create the orm
	orm := realorm.New(*DSN, database.PG)

	// create the post
	err := orm.Create(&Post{
		Title:   "Hello World",
		Content: "This is a test post",
	})

	if err != nil {
		panic(err)
	}

	// find all posts
	var posts []Post

	// findallpaginated
	var paginatedResult *realorm.PaginatedResult

	paginatedResult, err = orm.FindAllPaginated(&posts, 1, 10, nil)
	if err != nil {
		log.Fatalf("pagination error: %v\n", err)
	}

	fmt.Printf("paginatedResult: %+v\n", paginatedResult)

	// Print the posts
	err = orm.FindAll(&posts, nil)
	for _, post := range posts {
		fmt.Printf("%d, %s, %s\n", post.ID, post.Title, post.Content)

		// Update the post
		updated, err := orm.Update(Post{Title: fmt.Sprintf("Updated title: %d", post.ID)}, post.ID, &realorm.WhereClause{
			Query: "id = ?",
			Args:  []interface{}{post.ID}})

		if err != nil {
			panic(err)
		}

		updatedPost := updated.(*Post)
		fmt.Printf("%d, %s, %s\n", updatedPost.ID, updatedPost.Title, updatedPost.Content)

		// Delete the post
		err = orm.Delete(&Post{}, &realorm.WhereClause{
			Query: "id = ?",
			Args:  []interface{}{post.ID},
		})

		if err != nil {
			panic(err)
		}
	}
}
