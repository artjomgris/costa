package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pkg/browser"
	"golang.org/x/crypto/bcrypt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type Seller struct {
	Id, FName, LName, Role, Pass string
}
type Product struct {
	Name, Date, Expires string
	Id, Qnty            int
}

var UserAuth Seller

func logoff(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
	switch r.Method {
	case "GET":
		UserAuth = Seller{}
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func product(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
	switch r.Method {
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		qnty, _ := strconv.Atoi(r.FormValue("qnty"))
		product := Product{
			Name:    r.FormValue("name"),
			Qnty:    qnty,
			Date:    r.FormValue("date"),
			Expires: r.FormValue("before"),
		}

		if _, err := os.Stat("src/products.json"); errors.Is(err, os.ErrNotExist) {
			product.Id = 0
			enc_product, err := json.Marshal(product)
			if err != nil {
				panic(err)
			}
			db, err := os.Create("src/products.json")
			if err != nil {
				panic(err)
			}
			defer db.Close()
			db.Write([]byte(string(enc_product) + "\n"))

			w.WriteHeader(http.StatusOK)
		} else {
			db, err := ioutil.ReadFile("src/products.json")
			if err != nil {
				panic(err)
			}
			tempDb := strings.Split(string(db), "\n")
			var tmppr Product
			if len(tempDb) <= 1 {
				product.Id = 0
			} else {
				json.Unmarshal([]byte(tempDb[len(tempDb)-2]), &tmppr)
				product.Id = tmppr.Id + 1
			}

			enc_product, err := json.Marshal(product)
			if err != nil {
				panic(err)
			}

			f, err := os.OpenFile("src/products.json",
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Println(err)
			}
			defer f.Close()
			if _, err := f.WriteString(string(enc_product) + "\n"); err != nil {
				log.Println(err)
			}
			w.WriteHeader(http.StatusOK)
		}
	case "GET":
		db, err := ioutil.ReadFile("src/products.json")
		if err != nil {
			panic(err)
		}
		tempDb := strings.Split(string(db), "\n")
		var products []Product
		tempDb = tempDb[:len(tempDb)-1]
		for _, i := range tempDb {
			var tmp_product Product
			json.Unmarshal([]byte(i), &tmp_product)
			products = append(products, tmp_product)
		}
		marshalled, err := json.Marshal(products)
		if err != nil {
			panic(err)
		}
		w.Write(marshalled)
	case "PATCH":
		responseData, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		var tempDb []string
		if responseData[0] == 'u' {
			responseData = responseData[1:]
			if err := r.ParseForm(); err != nil {
				fmt.Fprintf(w, "ParseForm() err: %v", err)
				return
			}
			var res Product
			json.Unmarshal(responseData, &res)
			fmt.Println(res)
			db, err := ioutil.ReadFile("src/products.json")
			if err != nil {
				panic(err)
			}
			tempDb = strings.Split(string(db), "\n")
			tempDb = tempDb[:len(tempDb)-1]
			for j, i := range tempDb {
				var tmp_product Product
				json.Unmarshal([]byte(i), &tmp_product)
				if tmp_product.Id == res.Id {
					tmp_product = res
					marshalled, err := json.Marshal(tmp_product)
					if err != nil {
						panic(err)
					}
					tempDb[j] = string(marshalled)
				}
			}
		} else {
			ifDel := false
			var res struct {
				Id int
			}
			if responseData[0] == 'd' {
				ifDel = true
				responseData = responseData[1:]
			}
			json.Unmarshal(responseData, &res)
			db, err := ioutil.ReadFile("src/products.json")
			if err != nil {
				panic(err)
			}
			tempDb = strings.Split(string(db), "\n")
			tempDb = tempDb[:len(tempDb)-1]
			for j, i := range tempDb {
				var tmp_product Product
				json.Unmarshal([]byte(i), &tmp_product)
				if tmp_product.Id == res.Id {
					if !ifDel {
						tmp_product.Qnty -= 1
						marshalled, err := json.Marshal(tmp_product)
						if err != nil {
							panic(err)
						}
						tempDb[j] = string(marshalled)
					} else {
						if j == len(tempDb)-1 {
							tempDb = tempDb[:j]
						} else {
							tempDb = append(tempDb[:j], tempDb[j+1:]...)
						}

					}
				}
			}
		}
		var db_string string
		for _, i := range tempDb {
			db_string += i + "\n"
		}

		f, err := os.OpenFile("src/products.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			log.Fatal(err)
		}
		err = f.Truncate(0)
		_, err = f.Seek(0, 0)

		defer f.Close()
		f.Write([]byte(db_string))
		w.WriteHeader(http.StatusOK)

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func auth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if UserAuth.Id == "" {
			users, err := os.Open("src/users.json")
			defer users.Close()
			if err != nil && fmt.Sprint(err)[0:4] == "open" {
				http.ServeFile(w, r, "src/new.html")
			} else {
				http.ServeFile(w, r, "src/index.html")
			}
		} else {
			http.Redirect(w, r, "/panel", http.StatusFound)
		}
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		id := r.FormValue("seller")
		db, err := ioutil.ReadFile("src/users.json")
		if err != nil {
			panic(err)
		}
		tempDb := strings.Split(string(db), "\n")
		var user Seller
		for _, i := range tempDb {
			var tmp_user Seller
			json.Unmarshal([]byte(i), &tmp_user)
			if tmp_user.Id == id {
				user = tmp_user
			}
		}
		if user.Id == "" {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.Write([]byte(user.Pass))
		}
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func panel(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
	switch r.Method {
	case "GET":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if UserAuth.Id == "" {
			http.Redirect(w, r, "/", http.StatusForbidden)
		} else {
			t, err := template.ParseFiles("src/panel.html")
			if err != nil {
				panic(err)
			}
			data := struct {
				User    Seller
				IfAdmin bool
			}{
				User:    UserAuth,
				IfAdmin: (UserAuth.Role == "admin"),
			}
			err = t.Execute(w, data)
			if err != nil {
				panic(err)
			}
		}

	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		id := r.FormValue("seller")
		db, err := ioutil.ReadFile("src/users.json")
		if err != nil {
			panic(err)
		}
		tempDb := strings.Split(string(db), "\n")
		var user Seller
		for _, i := range tempDb {
			var tmp_user Seller
			json.Unmarshal([]byte(i), &tmp_user)
			if tmp_user.Id == id {
				user = tmp_user
			}
		}
		if user.Id == "" {
			w.WriteHeader(http.StatusForbidden)
		} else {
			w.Write([]byte(user.Pass))
		}
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func checkpass(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
	switch r.Method {
	case "GET":
		w.WriteHeader(404)
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		if CheckPasswordHash(r.FormValue("Pass"), r.FormValue("Hash")) {
			w.WriteHeader(http.StatusAccepted)
			UserAuth.Id = r.FormValue("Id")
			db, err := ioutil.ReadFile("src/users.json")
			if err != nil {
				panic(err)
			}
			tempDb := strings.Split(string(db), "\n")
			for _, i := range tempDb {
				var tmp_user Seller
				json.Unmarshal([]byte(i), &tmp_user)
				if tmp_user.Id == UserAuth.Id {
					UserAuth = tmp_user
					break
				}
			}
			w.Write([]byte("/panel"))
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

func user(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, private, max-age=0")
	switch r.Method {
	case "GET":
		db, err := ioutil.ReadFile("src/users.json")
		if err != nil {
			panic(err)
		}
		tempDb := strings.Split(string(db), "\n")
		var users []Seller
		tempDb = tempDb[:len(tempDb)-1]
		for _, i := range tempDb {
			var tmp_user Seller
			json.Unmarshal([]byte(i), &tmp_user)
			users = append(users, tmp_user)
		}
		marshalled, err := json.Marshal(users)
		if err != nil {
			panic(err)
		}
		w.Write(marshalled)
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		pass, err := HashPassword(r.FormValue("Pass"))
		if err != nil {
			panic(err)
		}
		user := Seller{
			FName: r.FormValue("FName"),
			LName: r.FormValue("LName"),
			Role:  r.FormValue("Role"),
			Pass:  pass,
		}
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		user.Id = string(user.FName[0]) + string(user.LName[0]) + strconv.Itoa(int(r1.Float64()*10000))

		users, err := os.Open("src/users.json")
		defer users.Close()
		var db *os.File
		if err != nil && fmt.Sprint(err)[0:4] == "open" {
			db, err = os.Create("src/users.json")
			if err != nil {
				panic(err)
			}
		} else {
			user.Role = "seller"
			db, err = os.OpenFile("src/users.json",
				os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Println(err)
			}
		}
		defer db.Close()

		enc_user, err := json.Marshal(user)
		if err != nil {
			panic(err)
		}
		db.Write([]byte(string(enc_user) + "\n"))
		w.Write([]byte(user.Id))
	case "PATCH":
		responseData, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		var user Seller
		err = json.Unmarshal(responseData, &user)
		if user.Id != UserAuth.Id {
			db, err := ioutil.ReadFile("src/users.json")
			if err != nil {
				panic(err)
			}
			tempDb := strings.Split(string(db), "\n")
			tempDb = tempDb[:len(tempDb)-1]
			var users string
			for _, i := range tempDb {
				var tmpuser Seller
				err := json.Unmarshal([]byte(i), &tmpuser)
				if err != nil {
					panic(err)
				}
				if tmpuser.Id == user.Id {
					tmpuser.Role = user.Role
				}
				tmpusrstr, err := json.Marshal(tmpuser)
				if err != nil {
					panic(err)
				}
				users += string(tmpusrstr) + "\n"
			}

			f, err := os.OpenFile("src/users.json",
				os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				log.Println(err)
			}
			defer f.Close()

			if err != nil {
				panic(err)
			}
			if _, err := f.WriteString(users); err != nil {
				log.Println(err)
			}
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusForbidden)
		}
	case "DELETE":
		responseData, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		var user Seller
		err = json.Unmarshal(responseData, &user)

		if user.Id != UserAuth.Id {
			db, err := ioutil.ReadFile("src/users.json")
			if err != nil {
				panic(err)
			}
			tempDb := strings.Split(string(db), "\n")
			tempDb = tempDb[:len(tempDb)-1]
			var users string
			for _, i := range tempDb {
				var tmpuser Seller
				err := json.Unmarshal([]byte(i), &tmpuser)
				if err != nil {
					panic(err)
				}
				if tmpuser.Id != user.Id {
					tmpusrstr, err := json.Marshal(tmpuser)
					if err != nil {
						panic(err)
					}
					users += string(tmpusrstr) + "\n"
				}

			}

			f, err := os.OpenFile("src/users.json",
				os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				log.Println(err)
			}
			defer f.Close()

			if err != nil {
				panic(err)
			}
			if _, err := f.WriteString(users); err != nil {
				log.Println(err)
			}
		} else {
			w.WriteHeader(http.StatusForbidden)
		}

	}

}

func main() {

	http.HandleFunc("/", auth)
	http.HandleFunc("/checkpass", checkpass)
	http.HandleFunc("/panel", panel)
	http.HandleFunc("/logoff", logoff)
	http.HandleFunc("/product", product)
	http.HandleFunc("/user", user)

	http.Handle("/src/css/", http.StripPrefix("/src/css/", http.FileServer(http.Dir("src/css/"))))
	http.Handle("/src/js/", http.StripPrefix("/src/js/", http.FileServer(http.Dir("src/js/"))))
	http.Handle("/src/", http.StripPrefix("/src/", http.FileServer(http.Dir("src/"))))
	browser.OpenURL("http://localhost:8090")
	http.ListenAndServe(":8090", nil)

}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
