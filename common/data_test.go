package common

import (
	"bytes"
	"encoding/gob"
	"github.com/kr/pretty"
	"testing"
)

type TestData struct {
	Vi  int
	Vf  float32
	Vai []int
	Vmi map[string]int
}

func Test_DataItem(t *testing.T) {
	t.Log(Norf("DataItem Start"))
	var item1 DataItem
	err := item1.Parse([]int{1, 2, 5, 7})
	if err != nil {
		t.Fatal(Errf("Parse1 Err:%v", err))
	}
	if item1.Version == 0 {
		t.Log(Infof("New Item1 Version OK"))
	}
	decoder := gob.NewDecoder(bytes.NewBuffer(item1.Data))
	var data1 []int
	err = decoder.Decode(&data1)
	if err != nil {
		t.Fatal(Errf("Decode Err:%v", err))
	}
	if data1[0] == 1 && data1[3] == 7 {
		t.Log(Infof("New Item1 Data OK:%v\nItem:%v", data1, item1))
	} else {
		t.Fatal(Errf("New Item1 Data Failed:%v\nItem:%v", data1, item1))
	}

	var item2 DataItem
	item1Bytes, err := item1.Bytes()
	if err != nil {
		t.Fatal(Errf("Bytes1 Err:%v", err))
	}
	err = item2.Parse(item1Bytes)
	if err != nil {
		t.Fatal(Errf("Parse2 Err:%v", err))
	}
	item2buf := bytes.NewBuffer(item2.Data)
	decoder = gob.NewDecoder(item2buf)
	var data2 []int
	err = decoder.Decode(&data2)
	diff12 := pretty.Diff(data2, data1)
	if len(diff12) <= 0 {
		t.Log(Infof("New Item2 Data OK:%v\nItem:%v\nDiff:%v\n", data2, item2, diff12))
	} else {
		t.Fatal(Errf("New Item2 Data Failed:%v\nItem:%v\nDiff:%v\n", data2, item2, diff12))
	}

	var item3 DataItem
	err = item3.Parse(item2)
	if err != nil {
		t.Fatal(Errf("Parse3 Err:%v", err))
	}
	diff23 := pretty.Diff(item3.Data, item2.Data)
	if len(diff23) <= 0 {
		t.Log(Infof("New Item3 OK:\nItem3:%v\nItem2:%v\nDiff:%v\n", item3, item2, diff23))
	} else {
		t.Fatal(Errf("New Item3 Failed:\nItem3:%v\nItem2:%v\nDiff:%v\n", item3, item2, diff23))
	}

	var item4 DataItem
	var data4 TestData
	data4.Vi = 5
	data4.Vf = 0.2
	data4.Vai = []int{1, 2, 6, 7}
	data4.Vmi = map[string]int{"a": 1, "b": 2, "hello": 3}
	err = item4.Parse(data4)
	if err != nil {
		t.Fatal(Errf("Parse4 Err:%v", err))
	}
	item4buf := bytes.NewBuffer(item4.Data)
	decoder = gob.NewDecoder(item4buf)
	var data4out TestData
	err = decoder.Decode(&data4out)
	if err != nil {
		t.Fatal(Errf("Decode4 Err:%v", err))
	}
	diff4 := pretty.Diff(data4, data4out)
	if len(diff4) <= 0 {
		t.Log(Infof("New Item4 OK:\nItem4:%v\nData4:%v\nDiff:%v\n", item4, data4out, diff4))
	} else {
		t.Fatal(Errf("New Item4 Failed:\nItem4:%v\nData4:%v\nDiff:%v\n", item4, data4out, diff4))
	}

	var item5 DataItem
	item4Bytes, err := item4.Bytes()
	if err != nil {
		t.Fatal(Errf("Bytes4 Err:%v", err))
	}
	err = item5.Parse(item4Bytes)
	if err != nil {
		t.Fatal(Errf("Parse5 Err:%v", err))
	}
	item5buf := bytes.NewBuffer(item5.Data)
	decoder = gob.NewDecoder(item5buf)
	var data5 TestData
	err = decoder.Decode(&data5)
	diff45 := pretty.Diff(data5, data4)
	if len(diff45) <= 0 {
		t.Log(Infof("New Item5 Data OK:\nData:%v\nItem:%v\nDiff:%v\n", data5, item5, diff45))
	} else {
		t.Fatal(Errf("New Item5 Data Failed:\nData:%v\nItem:%v\nDiff:%v\n", data5, item5, diff45))
	}
	t.Log(Norf("DataItem End"))
}

func Test_DataSet(t *testing.T) {

	set1 := new(BaseDataSet)
	var item1 DataItem
	data1 := TestData{
		Vi:  5,
		Vf:  0.5,
		Vai: []int{2, 3, 7, 8},
		Vmi: map[string]int{"Hello": 1},
	}
	item1.Parse(data1)
	ok := set1.Set("Hello", item1)
	if ok {
		t.Log(Infof("Set Hello OK\n%v", set1))
	} else {
		t.Fatal(Errf("Set Hello Failed\n%v", set1))
	}
	var item2 DataItem
	item2, ok = set1.Get("Hello")
	if ok {
		t.Log(Infof("Get Hello OK\n%v", set1))
	} else {
		t.Fatal(Errf("Get Hello Failed\n%v", set1))
	}
	diff12 := pretty.Diff(item1.Data, item2.Data)
	if len(diff12) <= 0 {
		t.Log(Infof("12 OK\n%v", set1))
	} else {
		t.Log(Infof("12 OK\nitem1\n%v\nitem2\n%v\ndiff\n%v", item1, item2, diff12))
	}

	item3 := set1.Grant("Hello", item2)
	ok = set1.Set("Hello", item3)
	if ok {
		t.Log(Infof("Set Hello OK\n%v", set1))
	} else {
		t.Fatal(Errf("Set Hello Failed\n%v", set1))
	}
	ok = set1.Reverse("Hello")
	if ok {
		t.Log(Infof("Reverse Hello OK\n%v", set1))
	} else {
		t.Fatal(Errf("Reverse Hello Failed\n%v", set1))
	}

	ok = set1.Lock("Hello")
	if ok && set1.IsLocked("Hello") {
		t.Log(Infof("Lock1 Hello OK\n%v", set1))
	} else {
		t.Fatal(Errf("Lock1 Hello Failed\n%v", set1))
	}

	ok = set1.Lock("Hello")
	if !ok {
		t.Log(Infof("Lock2 Hello Failed\n%v", set1))
	} else {
		t.Fatal(Errf("Lock2 Hello OK\n%v", set1))
	}

	set1.Unlock("Hello")
	if !set1.IsLocked("Hello") {
		t.Log(Infof("Lock3 Hello OK\n%v", set1))
	} else {
		t.Fatal(Errf("Lock3 Hello Failed\n%v", set1))
	}

	t.Log(Norf("DataSet End"))
}
