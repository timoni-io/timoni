package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync/atomic"
)

type Value struct {
	atomic.Int64
}

func (v *Value) MarshalJSON() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v *Value) String() string {
	return fmt.Sprint(v.Load())
}

type Avg struct {
	Num, Div *Value
}

func (v Avg) MarshalJSON() ([]byte, error) {
	return []byte(v.String()), nil
}

func (v Avg) String() string {
	x := float64(v.Div.Load())
	if x < 0 {
		return "-1"
	} else if x == 0 {
		return "0"
	} else {
		return fmt.Sprintf("%.2f", float64(v.Num.Load())/x)
	}
}

func Handler(metrics any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		buf, _ := json.MarshalIndent(metrics, "", "  ")
		w.Header().Set("Content-Type", "application/json")
		w.Write(buf)
	}
}

func HandlerPretty(metrics any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		html1 := []byte(`
<html>
	<body>
	  <font size="6">`)
		html2 := []byte(`
	</font>
</body>
</html>`)

		buf, err := json.MarshalIndent(metrics, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		buf = bytes.ReplaceAll(buf, []byte("},\n"), []byte(""))
		buf = bytes.ReplaceAll(buf, []byte("}"), []byte(""))
		buf = bytes.ReplaceAll(buf, []byte("{"), []byte(""))
		buf = bytes.ReplaceAll(buf, []byte("\n"), []byte("<br>"))
		buf = append(html1, buf...)
		buf = append(buf, html2...)
		w.WriteHeader(http.StatusOK)
		w.Write(buf)
	}
}

type System struct {
	HeapActive uint64
	HeapIdle   uint64
	Goro       uint64
	Cpu        uint16
}

func (s System) MarshalJSON() ([]byte, error) {
	return []byte(s.String()), nil
}

func (s System) String() string {
	mm := runtime.MemStats{}
	runtime.ReadMemStats(&mm)
	s.HeapActive = uint64(mm.HeapInuse) / (1 << 20)
	s.HeapIdle = uint64(mm.HeapIdle) / (1 << 20)
	s.Goro = uint64(runtime.NumGoroutine())
	return fmt.Sprintf(`{"HeapActive(MB)": %d,"HeapIdle(MB)": %d,"Goro": %d,"CPU": %d}`, s.HeapActive, s.HeapIdle, s.Goro, s.Cpu)
}
