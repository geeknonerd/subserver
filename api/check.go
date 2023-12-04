package handler
 
import (
  "fmt"
  "net/http"
)
 
func Check(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "OK!")
}
