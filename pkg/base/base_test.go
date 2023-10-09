package base

import (
	"fmt"
	"os"
	"testing"

	boltdb "among-us-roombot/pkg/base/boltdb"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/boltdb/bolt"
)

func TestGet(t *testing.T) {

	db, err := bolt.Open("bolt-test.db", 0600, nil)
	if err != nil {
		fmt.Println(err)
	}

	base := boltdb.NewBase(db)

	base.Save("key", "value", "bucket")
	base.Save("ключ", "значение", "корзина")

	t.Cleanup(func() {
		db.Close()
		os.Remove("bolt-test.db")
	})

	Convey("Тест нормального чтения", t, func() {
		value, err := base.Get("key", "bucket")
		if err != nil {
			fmt.Println(err)
		}
		So("value", ShouldEqual, value)
	})
	Convey("Тест чтения несуществующего ключя", t, func() {
		value, err := base.Get("null", "bucket")
		if err != nil {
			fmt.Println(err)
		}
		So("", ShouldEqual, value)
	})
	Convey("Тест чтения несуществующего bucket", t, func() {
		value, err := base.Get("key", "null")
		if err != nil {
			fmt.Println(err)
		}
		So("", ShouldEqual, value)
	})
	Convey("Тест кириллицы", t, func() {
		value, err := base.Get("ключ", "корзина")
		if err != nil {
			fmt.Println(err)
		}
		So("значение", ShouldEqual, value)
	})
}

func TestSave(t *testing.T) {
	db, err := bolt.Open("bolt-test.db", 0600, nil)
	if err != nil {
		fmt.Println(err)
	}

	base := boltdb.NewBase(db)

	t.Cleanup(func() {
		db.Close()
		os.Remove("bolt-test.db")
	})

	Convey("Тест нормального сохранения", t, func() {
		base.Save("key", "value", "bucket")
		value, err := base.Get("key", "bucket")
		if err != nil {
			fmt.Println(err)
		}
		So("value", ShouldEqual, value)
	})
	Convey("Тест сохранения спецсимволов", t, func() {
		base.Save("key", "\\\\", "bucket")
		value, err := base.Get("key", "bucket")
		if err != nil {
			fmt.Println(err)
		}
		So("\\\\", ShouldEqual, value)
	})
	Convey("Тест сохранения спецсимволов", t, func() {
		base.Save("key", "/n", "bucket")
		value, err := base.Get("key", "bucket")
		if err != nil {
			fmt.Println(err)
		}
		So("/n", ShouldEqual, value)
	})
	Convey("Тест сохранения спецсимволов", t, func() {
		base.Save("key", ".....", "bucket")
		value, err := base.Get("key", "bucket")
		if err != nil {
			fmt.Println(err)
		}
		So(".....", ShouldEqual, value)
	})
	Convey("Тест сохранения пустого значения", t, func() {
		base.Save("key", "", "bucket")
		value, err := base.Get("key", "bucket")
		if err != nil {
			fmt.Println(err)
		}
		So("", ShouldEqual, value)
	})
	Convey("Тест сохранения пустого ключа", t, func() {
		base.Save("", "value", "bucket")
		value, err := base.Get("", "bucket")
		if err != nil {
			fmt.Println(err)
		}
		So("", ShouldEqual, value)
	})
	Convey("Тест сохранения в пустой bucket", t, func() {
		base.Save("key", "value", "")
		value, err := base.Get("key", "bucket")
		if err != nil {
			fmt.Println(err)
		}
		So("", ShouldEqual, value)
	})
}
