package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"go-rest-api/hello"
)

// ------------------------- Modèles ----------------------------------

// La structure Question représente une question avec un identifiant unique et une chaîne de caractères pour la question.
type Question struct {
	ID            string `json:"id"`
	Question      string `json:"question"`
	Answer1       string `json:"answer1"`
	Answer2       string `json:"answer2"`
	Answer3       string `json:"answer3"`
	Answer4       string `json:"answer4"`
	CorrectAnswer int    `json:"correctAnswer"`
}

// La variable "questions" est une tranche ("slice") de type "Question", qui sert à stocker toutes les questions.
// En d'autres termes, il s'agit d'un tableau dynamique capable de stocker un nombre variable d'éléments de type "Question".
var questions []Question

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

var users []User

// use godot package to load/read the .env file and
// return the value of the key
func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

// vérification des erreurs. "panic" met fin à l'exécution du code.
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// connection à la base de données
func setupDB() *sql.DB {
	DB_NAME := goDotEnvVariable("DB_NAME")
	DB_USER := goDotEnvVariable("DB_USER")
	DB_PASSWORD := goDotEnvVariable("DB_PASSWORD")

	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
	DB, err := sql.Open("postgres", dbinfo)

	checkErr(err)

	return DB
}

// Struct pour la réponse reçue au format json
type JsonResponse struct {
	Type    string     `json:"type"`
	Data    []Question `json:"data"`
	Message string     `json:"message"`
}

// ------------------- Contrôleurs ------------------------------

// renvoie une string quand on se connecte sur le endpoint /
func homeLink(w http.ResponseWriter, r *http.Request) {
	//dotenv := goDotEnvVariable("DBNAME")
	fmt.Fprintf(w, "Hello!")
}

// La fonction getQuestions renvoie toutes les questions sous forme de JSON.
func getQuestions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(questions)
}

// La fonction deleteQuestion supprime une question spécifique de la slice des questions.
func deleteQuestion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for index, item := range questions {
		if item.ID == params["id"] {
			questions = append(questions[:index], questions[index+1:]...)
			break
		}
	}
	json.NewEncoder(w).Encode(questions)
}

// La fonction getQuestion renvoie une question spécifique par son ID.
func getQuestion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	for _, item := range questions {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			break
		}
	}
}

// La fonction createQuestion crée une nouvelle question à partir des données JSON de la requête HTTP POST.
func createQuestion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var question Question
	_ = json.NewDecoder(r.Body).Decode(&question)
	question.ID = strconv.Itoa(rand.Intn(10000000))
	questions = append(questions, question)
	json.NewEncoder(w).Encode(questions)
}

// La fonction updateQuestion met à jour une question existante avec les données JSON de la requête HTTP PUT.
func updateQuestion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	var question Question
	_ = json.NewDecoder(r.Body).Decode(&question)
	for index, item := range questions {
		if item.ID == params["id"] {
			questions = append(questions[:index], questions[index+1:]...)
			question.ID = params["id"]
			questions = append(questions, question)
			break
		}
	}
	json.NewEncoder(w).Encode(questions)
}

// Function for handling messages
func printMessage(message string) {
	fmt.Println("")
	fmt.Println(message)
	fmt.Println("")
}

// --------------------- contrôleurs BDD ------------------------------

// ------------------------- Questions ---------------------------------
// get all questions
func getQuestionsSQL(w http.ResponseWriter, r *http.Request) {
	db := setupDB()
	printMessage("Getting questions...")
	// Récupérer toutes les questions de la table "questions"
	rows, err := db.Query("Select * FROM questions")
	// Vérifier les erreurs
	checkErr(err)
	var questions []Question
	// Parcourir les questions
	for rows.Next() {
		var id string
		var questionText string
		var answer1 string
		var answer2 string
		var answer3 string
		var answer4 string
		var correctAnswer int
		err = rows.Scan(&id, &questionText, &answer1, &answer2, &answer3, &answer4, &correctAnswer)
		// Vérifier les erreurs
		checkErr(err)
		question := Question{
			ID:            id,
			Question:      questionText,
			Answer1:       answer1,
			Answer2:       answer2,
			Answer3:       answer3,
			Answer4:       answer4,
			CorrectAnswer: correctAnswer,
		}
		questions = append(questions, question)
	}
	var response = JsonResponse{
		Type:    "success",
		Data:    questions,
		Message: "Questions retrieved successfully!",
	}
	json.NewEncoder(w).Encode(response)
}

// Accède à une question
func getQuestionSQL(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	questionID := params["id"]

	var response = JsonResponse{}
	var question Question

	if questionID == "" {
		response = JsonResponse{Type: "error", Message: "Missing question ID."}
	} else {
		db := setupDB()

		printMessage("Getting question from DB")

		retrieved, err := db.Query("SELECT * from questions WHERE ID=$1", questionID)

		checkErr(err)

		for retrieved.Next() {
			var id string
			var questionText string
			var answer1 string
			var answer2 string
			var answer3 string
			var answer4 string
			var correctAnswer int
			err = retrieved.Scan(&id, &questionText, &answer1, &answer2, &answer3, &answer4, &correctAnswer)
			// Vérifier les erreurs
			checkErr(err)
			question.ID = id
			question.Question = questionText
			question.Answer1 = answer1
			question.Answer2 = answer2
			question.Answer3 = answer3
			question.Answer4 = answer4
			question.CorrectAnswer = correctAnswer

			var response = JsonResponse{
				Type:    "success",
				Data:    []Question{question},
				Message: "Question retrieved successfully!",
			}
			json.NewEncoder(w).Encode(response)
		}

		json.NewEncoder(w).Encode(response)
	}
}

// Crée une question
func createQuestionSQL(w http.ResponseWriter, r *http.Request) {
	var question Question
	_ = json.NewDecoder(r.Body).Decode(&question)

	printMessage("hello" + question.Question)

	var response = JsonResponse{}

	if question.Question == "" || question.Answer1 == "" || question.Answer2 == "" || question.Answer3 == "" || question.Answer4 == "" {
		response = JsonResponse{Type: "error", Message: "Missing data."}
	} else {
		db := setupDB()

		printMessage("Creating question into DB")

		fmt.Println("Inserting new question : " + question.Question)

		var lastInsertID int
		err := db.QueryRow("INSERT INTO questions (question, answer1, answer2, answer3, answer4, correct_answer) VALUES ($1, $2, $3, $4, $5, $6) returning id;", question.Question, question.Answer1, question.Answer2, question.Answer3, question.Answer4, question.CorrectAnswer).Scan(&lastInsertID)

		// check errors
		checkErr(err)

		response = JsonResponse{Type: "success", Message: "The question has been inserted successfully!"}
	}

	json.NewEncoder(w).Encode(response)
}

// Supprime une question
func deleteQuestionSQL(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	questionID := params["id"]

	var response = JsonResponse{}

	if questionID == "" {
		response = JsonResponse{Type: "error", Message: "Missing question ID."}
	} else {
		db := setupDB()

		printMessage("Deleting question from DB")

		_, err := db.Exec("DELETE FROM questions where ID = $1", questionID)

		// check errors
		checkErr(err)

		response = JsonResponse{Type: "success", Message: "The question has been deleted successfully!"}
	}

	json.NewEncoder(w).Encode(response)
}

// Modifie une question
func updateQuestionSQL(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	questionID := params["id"]

	var question Question
	_ = json.NewDecoder(r.Body).Decode(&question)

	var response = JsonResponse{}

	if questionID == "" {
		response = JsonResponse{Type: "error", Message: "Missing question ID."}
	} else {
		db := setupDB()

		printMessage("Updating question from DB")

		_, err := db.Exec("UPDATE questions SET question=$1, answer1=$2, answer2=$3, answer3=$4, answer4=$5, correct_answer=$6 WHERE ID = $7", question.Question, question.Answer1, question.Answer2, question.Answer3, question.Answer4, question.CorrectAnswer, questionID)

		// check errors
		checkErr(err)

		response = JsonResponse{Type: "success", Message: "The question has been updated successfully : " + question.Question}
	}

	json.NewEncoder(w).Encode(response)
}

//------------------------- Users ---------------------------------

func getUsers(w http.ResponseWriter, r *http.Request) {
	db := setupDB()
	printMessage("Getting users...")
	rows, err := db.Query("SELECT id, name, email, password FROM users")
	checkErr(err)
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password)
		checkErr(err)
		users = append(users, user)
	}

	err = rows.Err()
	checkErr(err)

	json.NewEncoder(w).Encode(users)
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	_ = json.NewDecoder(r.Body).Decode(&user)

	printMessage("hello" + "" + user.Name)

	var response = JsonResponse{}

	if user.Name == "" || user.Email == "" || user.Password == "" {
		response = JsonResponse{Type: "error", Message: "Missing data."}
	} else {
		db := setupDB()

		printMessage("Creating user into DB")

		fmt.Println("Inserting new user : " + user.Name)

		var lastInsertID int
		err := db.QueryRow("INSERT INTO users (name, email, password) VALUES ($1, $2, $3) returning id;", user.Name, user.Email, user.Password).Scan(&lastInsertID)

		// check errors
		checkErr(err)

		response = JsonResponse{Type: "success", Message: "The user has been inserted successfully!"}
	}

	json.NewEncoder(w).Encode(response)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	db := setupDB()
	params := mux.Vars(r)
	userID := params["id"]

	var user User
	err := db.QueryRow("SELECT id, name, email, password FROM users WHERE id = $1", userID).Scan(&user.ID, &user.Name, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
		} else {
			log.Fatal(err)
		}
		return
	}

	json.NewEncoder(w).Encode(user)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	checkErr(err)

	var response = JsonResponse{}

	if userID == "" {
		response = JsonResponse{Type: "error", Message: "Missing user ID."}
	} else {
		db := setupDB()

		printMessage("Updating user from DB")

		// Récupérer les données de l'utilisateur existant

		var existingUser User
		err := db.QueryRow("SELECT id, name, email, password FROM users WHERE id = $1", userID).Scan(&existingUser.ID, &existingUser.Name, &existingUser.Email, &existingUser.Password)

		checkErr(err)

		// Mettre à jour uniquement les champs non vides ou non nuls
		if user.Name == "" {
			if existingUser.Name != "" {
				user.Name = existingUser.Name
			}
		}
		if user.Email == "" {
			if existingUser.Email != "" {
				user.Email = existingUser.Email
			}
		}
		if user.Password == "" {
			if existingUser.Password != "" {
				user.Password = existingUser.Password
			}
		}

		_, err = db.Exec("UPDATE users SET name = $1, email = $2, password = $3 WHERE id = $4", user.Name, user.Email, user.Password, userID)
		checkErr(err)

		response = JsonResponse{Type: "success", Message: "User updated successfully!"}

		w.WriteHeader(http.StatusOK)
	}

	json.NewEncoder(w).Encode(response)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID := params["id"]
	var response = JsonResponse{}

	if userID == "" {
		response = JsonResponse{Type: "error", Message: "Missing user ID."}
	} else {
		db := setupDB()
		printMessage("Deleting user from DB")

		_, err := db.Exec("DELETE FROM users where ID = $1", userID)

		// check errors
		checkErr(err)

		response = JsonResponse{Type: "success", Message: "The user has been deleted successfully!"}
	}

	json.NewEncoder(w).Encode(response)
}

// ------------------------- MAIN et routes -------------------------------

func main() {
	// Initialise le routeur Mux et ajoute deux questions pour les tests.
	router := mux.NewRouter()
	questions = append(questions, Question{ID: "1", Question: "First question"})
	questions = append(questions, Question{ID: "2", Question: "Second question"})

	// Ajoute les endpoints pour les différentes requêtes HTTP.
	router.HandleFunc("/", homeLink).Methods("GET")

	router.HandleFunc("/questions", getQuestionsSQL).Methods("GET")
	router.HandleFunc("/questions/{id}", getQuestionSQL).Methods("GET")
	router.HandleFunc("/questions", createQuestionSQL).Methods("POST")
	router.HandleFunc("/questions/{id}", updateQuestionSQL).Methods("PUT")
	router.HandleFunc("/questions/{id}", deleteQuestionSQL).Methods("DELETE")

	router.HandleFunc("/users", getUsers).Methods("GET")
	router.HandleFunc("/users", createUser).Methods("POST")
	router.HandleFunc("/users/{id}", getUser).Methods("GET")
	router.HandleFunc("/users/{id}", deleteUser).Methods("DELETE")
	router.HandleFunc("/users/{id}", updateUser).Methods("PUT")

	fmt.Printf("Starting server at port 8085\n")
	log.Fatal(http.ListenAndServe(":8085", router))
}
