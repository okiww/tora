package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"okkybudiman/config"
	"okkybudiman/data"
	dataModel "okkybudiman/data/model"
	"os"
	"os/signal"
	"time"

	"okkybudiman/module/admin"
	"okkybudiman/module/user"
	u "okkybudiman/utility"

	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"golang.org/x/crypto/bcrypt"
)

var (
	appName = "API TRY OUT RUANG GURU"
	version = "development"

	runMigration    bool
	runSeeder       bool
	configuration   config.Configuration
	dbFactory       *data.DBFactory
	adminController *admin.Controller
	userController  *user.Controller
)
var identityKey = "id"

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type User struct {
	UserName string
	Email    string
}

func init() {
	//flag for migration and seeder if set true then running migration and seeder
	flag.BoolVar(&runMigration, "migrate", true, "run db migration before starting the server")
	flag.BoolVar(&runSeeder, "seeder", false, "run db seeder before starting the server")
	flag.Parse()

	cfg, err := config.New()
	if err != nil {
		glog.Fatalf("Failed to load configuration: %s", err)
		panic(fmt.Errorf("Fatal error loading configuration: %s", err))
	}

	configuration = *cfg
	dbFactory = data.NewDbFactory(configuration.Database)

	if runMigration {
		runDBMigration()
	}

	//inject dbFactory to admin controller
	adminController, err = admin.NewController(dbFactory)
	if err != nil {
		glog.Fatal(err.Error())
		panic(fmt.Errorf("Fatal error: %s", err))
	}

	//inject dbFactory to user controller
	userController, err = user.NewController(dbFactory)
	if err != nil {
		glog.Fatal(err.Error())
		panic(fmt.Errorf("Fatal error: %s", err))
	}
}

func setupRouter() *gin.Engine {
	glog.V(2).Info("Setting up server side routing")
	port := os.Getenv("PORT")
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	if port == "" {
		port = "8000"
	}
	// router.Use(cors.New(cors.Config{
	// 	AllowAllOrigins: true,
	// 	AllowMethods:    []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
	// 	AllowHeaders:    []string{"Origin", "Authorization", "Content-Type", "Access-Control-Allow-Origin"},
	// 	ExposeHeaders:   []string{"Content-Length"},
	// }))

	// the jwt middleware
	authMiddleware := jwt.GinJWTMiddleware{
		Realm:      "test zone",
		Key:        []byte("secret key"),
		Timeout:    time.Hour,
		MaxRefresh: time.Hour,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*User); ok {
				return jwt.MapClaims{
					identityKey: v.UserName,
				}
			}
			return jwt.MapClaims{}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals login
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			email := loginVals.Username
			password := loginVals.Password

			db, err := dbFactory.DBConnection()
			if err != nil {
				glog.Fatalf("Failed to open database connection: %s", err)
				panic(fmt.Errorf("Fatal error connecting to database: %s", err))
			}
			defer db.Close()

			var user dataModel.User

			if err := db.Where("email = ?", email).Find(&user).Error; err == nil {
				fmt.Println(user.Email)
				match := u.CheckPasswordHash(password, user.Password)

				if match {
					return &User{
						UserName: user.Name,
						Email:    user.Email,
					}, nil
				}
			}
			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header: Authorization, query: token, cookie: jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
	}

	router.POST("/login", authMiddleware.LoginHandler)

	router.NoRoute(authMiddleware.MiddlewareFunc(), func(c *gin.Context) {
		claims := jwt.ExtractClaims(c)
		log.Printf("NoRoute claims: %#v\n", claims)
		c.JSON(404, gin.H{"code": "PAGE_NOT_FOUND", "message": "Page not found"})
	})

	auth := router.Group("/auth")
	// Refresh time can be longer than token timeout
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)

	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware.MiddlewareFunc())
	{
		v1.GET("/hello", helloHandler)
		//api user
		user := v1.Group("/user")
		user.Use(CheckUser)
		{
			user.POST("/answer", userController.AnswerTest)
			user.POST("/attempt-test", userController.AttempTest)
			user.GET("/test/:id/result", userController.Result)
		}
		//api admin
		v1.Use(CheckAdmin)
		{
			v1.GET("/list-test", adminController.GetListTest)
			v1.GET("/test/:id/detail", adminController.GetDetailTest)

			v1.POST("/create-test", adminController.CreateTest)
			v1.POST("/create-question", adminController.CreateQuestion)
			v1.POST("/update-test", adminController.UpdateTest)
			v1.POST("/update-question", adminController.UpdateQuestion)
			v1.POST("/update-choice", adminController.UpdateChoice)

			v1.DELETE("/delete", adminController.DeleteTest)
			v1.DELETE("/delete-question", adminController.DeleteQuestion)
			v1.DELETE("/delete-choice", adminController.DeleteChoice)
		}

	}

	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}

	return router
}

func main() {
	r := setupRouter()

	srv := &http.Server{
		Addr:    configuration.Server.Port,
		Handler: r,
	}
	// Listen and Serve in 0.0.0.0:8080
	go func() {
		glog.Infof("Starting %s server version %s at %s", appName, version, configuration.Server.Port)
		if err := srv.ListenAndServe(); err != nil {
			glog.Fatalf("Failed to start server: %s", err)
			panic(fmt.Errorf("Fatal error failed to start server: %s", err))
		}
	}()

	// wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	glog.Info("Server shutted down")
}

func runDBMigration() {
	glog.Info("Running db migration")
	db, err := dbFactory.DBConnection()
	if err != nil {
		glog.Fatalf("Failed to open database connection: %s", err)
		panic(fmt.Errorf("Fatal error connecting to database: %s", err))
	}
	defer db.Close()

	db.AutoMigrate(
		&dataModel.User{},
		&dataModel.Role{},
		&dataModel.Test{},
		&dataModel.Question{},
		&dataModel.QuestionChoice{},
		&dataModel.UserAttemptTest{},
		&dataModel.UserAnswer{},
		&dataModel.UserScore{},
	)
	glog.Info("Done running db migration")

	if runSeeder {
		glog.Info("Running db seeder")
		var count int
		var admin_role uint
		var user_role uint
		db.Model(&dataModel.Role{}).Count(&count)
		if count == 0 {
			glog.V(1).Info("Running db seeder for table Currency")
			role1 := dataModel.Role{
				Name: "Admin",
			}
			db.Create(&role1)

			admin_role = role1.ID

			role2 := dataModel.Role{
				Name: "User",
			}
			db.Create(&role2)
			user_role = role2.ID
		}
		glog.V(1).Info("Running db seeder for table Currency")

		db.Model(&dataModel.User{}).Count(&count)
		if count == 0 {
			password := []byte("12345678")
			// Hashing the password with the default cost of 10
			hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
			if err != nil {
				panic(err)
			}

			user := dataModel.User{
				Name:     "Admin",
				Email:    "admin@admin.com",
				Password: string(hashedPassword),
				RoleID:   admin_role,
			}
			db.Create(&user)

			user2 := dataModel.User{
				Name:     "User",
				Email:    "user@user.com",
				Password: string(hashedPassword),
				RoleID:   user_role,
			}
			db.Create(&user2)
		}
	}
}

func CheckAdmin(c *gin.Context) {
	db, err := dbFactory.DBConnection()
	if err != nil {
		glog.Fatalf("Failed to open database connection: %s", err)
		panic(fmt.Errorf("Fatal error connecting to database: %s", err))
	}
	defer db.Close()

	claims := jwt.ExtractClaims(c)
	name := claims["id"].(string)

	var user dataModel.User
	var role dataModel.Role
	db.Where("name = ?", name).Find(&user)
	db.First(&role, user.RoleID)
	if role.Name != "Admin" {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "you cannot have access",
		})
		c.Abort()
	}

	c.Next()
}

func CheckUser(c *gin.Context) {
	db, err := dbFactory.DBConnection()
	if err != nil {
		glog.Fatalf("Failed to open database connection: %s", err)
		panic(fmt.Errorf("Fatal error connecting to database: %s", err))
	}
	defer db.Close()

	claims := jwt.ExtractClaims(c)
	name := claims["id"].(string)

	var user dataModel.User
	var role dataModel.Role
	db.Where("name = ?", name).Find(&user)
	db.First(&role, user.RoleID)
	if role.Name != "User" {
		c.JSON(400, gin.H{
			"code":    400,
			"message": "you cannot have access",
		})
		c.Abort()
	}

	c.Next()
}

func helloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	fmt.Println(claims["id"].(string))
	c.JSON(200, gin.H{
		"userID":   claims["id"],
		"userName": claims["id"].(string),
		"text":     "Hello World.",
	})
}
