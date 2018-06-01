package orm

import (
	"github.com/LuisZhou/lpge/orm"
	"github.com/jinzhu/gorm"
	"testing"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func TestDb(t *testing.T) {
	conns := make(map[string]string)
	conns["mysql"] = "www_db_admin:tankgogoGo123@(192.168.163.133)/ww?charset=utf8&parseTime=true&loc=Local"

	orm.Connect(conns, "mysql")

	db := orm.Default()

	// Migrate the schema
	db.AutoMigrate(&Product{})

	// Create
	db.Create(&Product{Code: "L1212", Price: 1000})

	// Read
	var product Product
	db.First(&product, 1)                   // find product with id 1
	db.First(&product, "code = ?", "L1212") // find product with code l1212

	// Update - update product's price to 2000
	var p uint = 2000
	db.Model(&product).Update("Price", p)

	// Delete - delete product
	db.Delete(&product)

	// drop table
	db.Exec("DROP TABLE products;")

	orm.Close()
}
