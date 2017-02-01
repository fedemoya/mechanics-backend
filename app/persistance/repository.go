package persistance

import (
    "strconv"
    _ "github.com/mattn/go-sqlite3"
    "github.com/jmoiron/sqlx"
    "database/sql"
    "log"
    "reflect"
)

type Repository struct {
    db *sqlx.DB
}

func NewRepository (dbName string) *Repository  {
    sqlx.NameMapper = func(s string) string { return s }
    // this Pings the database trying to connect, panics on error
    // use sqlx.Open() for sql.Open() semantics
    db, err := sqlx.Connect("sqlite3", "/data/" + dbName)
    if err != nil {
        log.Fatalln(err)
    }
    return &Repository{db}
}

func (r *Repository) Save (object interface{}) (int64, error) {

    v := reflect.ValueOf(object).Elem()
    t := v.Type()

    objectName := t.Name()

    var insert = "INSERT INTO " + objectName + " ("

    insert = insert +  t.Field(0).Name

    for i := 1; i < t.NumField(); i++ {
        insert = insert + ", " +  t.Field(i).Name
    }

    insert  = insert + ") VALUES (null, "

    numField := t.NumField()
    queryArgs := make([]interface{}, 0)

    // skip Field 0 because it is the Id and
    // it is generated by the database
    for i := 1; i < numField; i++ {

        field := v.Field(i)
        fieldKind := field.Kind()
        if (fieldKind.String() == "int64") {
            int64Field := field.Interface().(int64)
            queryArgs = append(queryArgs, int64Field)
        } else if (fieldKind.String() == "int") {
            intField := field.Interface().(int)
            queryArgs = append(queryArgs, intField)
        } else if (fieldKind.String() == "string") {
            stringField := field.Interface().(string)
            queryArgs = append(queryArgs, stringField)
        } else {
            panic("Unexpected field kind: " + fieldKind.String());
        }

        insert = insert + "?"
        if (i < numField - 1) {
            insert = insert + ", "
        }
    }

    insert = insert + ")"

    log.Println("About to exec the following sql:\n" + insert)
    log.Printf("queryArgs content: %v\n", queryArgs)

    var err error
    var res sql.Result
    res, err = r.db.Exec(insert, queryArgs...)

    if err != nil {
        return 0, err
    }

    var id int64
    id, err = res.LastInsertId()

    return id, err
}

func (r *Repository) Retrieve (emptyObject interface{}, id int64) error {

    t := reflect.TypeOf(emptyObject).Elem()

    objectName := t.Name()

    selectStr := "SELECT * FROM " + objectName + " WHERE Id = ?"

    err := r.db.Get(emptyObject, selectStr, strconv.FormatInt(id, 10));

    return err
}
