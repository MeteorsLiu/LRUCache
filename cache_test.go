package cache

import (
	"testing"
	"time"
)

func TestFunc(t *testing.T) {
	ce := New(2)
	ce.Set("test1", 114514)
	ce.SetWithExpire("test2", "2222fdg", 5*time.Second)
	time.Sleep(6 * time.Second)

	t.Log(ce.Get("test1"))
	t.Log(ce.GetAndRemoveExpire("test2"))
	t.Log(ce.GetAndRemoveExpire("test2"))
	ce.Set("test999", 114514)
	ce.Set("test3423", 114514)
	t.Log(ce.Get("test999"))
	t.Log(ce.Get("test3423"))
	// ce.Set("test1", 114564)
	t.Log(ce.Get("test1"))

	//Test Map
	var tmap = map[string]interface{}{}
	tmap["dsfsdf"] = 121223
	tmap["rere"] = "dsfsdfsd"
	ce.Set("testmap", tmap)

	TMAP, _ := ce.Get("testmap")
	t.Log(TMAP.(map[string]interface{})["rere"])
	t.Log(ce.Has("testmap"))

	t.Log(ce.Has("testmap"))
	ce.Remove("testmap")
	t.Log(ce.Has("testmap"))

	ce.Reset()

	ce.Set("test1", 114234)
	ce.SetWithExpire("test2", "2321322fdg", 5*time.Second)

	t.Log(ce.Get("test1"))

	t.Log(ce.GetAndRemoveExpire("test2"))

	ce.Clear()
}

func BenchMark(b *testing.B) {
	//b.ResetTimer()
	ce := New(114514)
	for i := 0; i < b.N; i++ {
		ce.Set(i, "test213123")
		ce.Get(i)
	}
	ce.Clear()

}
