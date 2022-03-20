# REALORM

**realorm** is a small wrapper around the [GORM](https://gorm.io) written in go.

It aims at providing a smooth interface to CRUD operations in a small to medium API.

## Basic functionality

- Unified way of connecting to the database
- A simple interface for implementing Find, FindAll, FindAllPaginated, Update, Delete operations.
- Handles pagination using page numbers and limit offset.
- Fully tested with 100% code coverage.
- Support for postgres, mysql, sqlite3

### Installation

```bash
go get github.com/abiiranathan/realorm
```

### Example Usage

```go

import (
  "github.com/abiiranathan/realorm/realorm"
)

type User struct{
  FirstName string `json:"first_name" gorm:"not null;" binding:"required"`
  LastName string `json:"last_name" gorm:"not null;" binding:"required"`
  Password string `json:"password" gorm:"not null;" binding:"required"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error{
  hash, err := auth.HashPassword(u.Password)

  if err != nil{
    u.Password = hash
  }

  // Returning an error will rollback gorm transaction
  return err
}

```

### Connect to database

```go
// initialize the ORM with the db
orm := realorm.New(dsn, database.PG)

```

### CREATE

```go

  user := &User{
    FirstName: "John",
    LastName: "Doe",
    Password: "password"
  }

  db.AutoMigrate(&User{})

  // pass a pinter to user model/entity
  err := orm.Create(user)
```

### UPDATE

```go

  updates := User{FirstName: "Jane"}

  // pass a user model/entity not a pointer
  // only none-zero values in updates are updated as in GORM
  updatedUser := orm.Update(
    updates, user.ID, &realorm.WhereClause{
      Query: "id = ?",
      Args:  []interface{}{user.ID}
      })

```

### FIND

```go
  // get a single record
  user := orm.Find(&User{}, &realorm.WhereClause{
      Query: "last_name LIKE ?",
      Args:  []interface{}{"Doe"}
      })
  
  // OR
  user := orm.Find(&User{}, &realorm.WhereClause{
      Query: "id = ? AND first_name=?",
      Args:  []interface{}{1, "Jane"}
      })

```

### FIND ALL RECORDS

```go
  // get all the records without filtering.
  // Pass where conditions for filtering records.
  users := orm.FindAll(&User{}, nil)

```

### Paginate records

```go
  
  var users []User
  var paginatedResult *realorm.PaginatedResult

  page := 1
  pageSize := 25

  paginatedResult, err := orm.FindAllPaginated(&users, page, pageSize, nil)

  // Paginated Result is an interface of form
  for _, user := range paginatedResult.Results{
    ...
  }

```

### DELETE  

```go

err := orm.Delete(&User{}, &realorm.WhereClause{
  Query: "id = ?", 
  Args:  []interface{}{1}
  })

```

### Advanced Usage

As you can tell, realorm is very small with limited but on point functionality.

For advanced use cases, get a reference to the underlying database
connection and use it as you please.

```go

db := orm.GetDB()

// Forexample as per GORM DOCs
// https://gorm.io/docs/preload.html
db.Preload("Orders").Preload("Profile").Preload("Role").Find(&users)

```

### Running the tests

go test ./...
