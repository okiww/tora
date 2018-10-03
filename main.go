package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"okkybudiman/config"
	"okkybudiman/data"
	dataModel "okkybudiman/data/model"
	u "okkybudiman/utility"
	"os"
	"os/signal"
	"time"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

var (
	appName = "API TRY OUT RUANG GURU"
	version = "development"

	runMigration  bool
	runSeeder     bool
	configuration config.Configuration
	dbFactory     *data.DBFactory
)

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
	router.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"PUT", "PATCH", "GET", "POST", "DELETE"},
		AllowHeaders:    []string{"Origin", "Authorization", "Content-Type", "Access-Control-Allow-Origin"},
		ExposeHeaders:   []string{"Content-Length"},
	}))

	// the jwt middleware
	authMiddleware := &jwt.GinJWTMiddleware{
		Realm:      "test zone",
		Key:        []byte("secret key"),
		Timeout:    time.Hour,
		MaxRefresh: time.Hour,
		Authenticator: func(c *gin.Context) (interface{}, error) {
			fmt.Println("kadieu")
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
			if v, ok := data.(*User); ok && v.UserName != "" {
				return true
			}

			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header:Authorization",
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
		v1.Use(CheckAdmin)
		{
			v1.GET("/admin", helloHandler)
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
		var id uint
		user := dataModel.User{}
		db.Model(&dataModel.Role{}).Count(&count)
		if count == 0 {
			glog.V(1).Info("Running db seeder for table Currency")
			role := dataModel.Role{
				Name: "Admin",
			}
			db.Create(&role)
			id = role.ID
		}
		glog.V(1).Info("Running db seeder for table Currency")

		if err := db.Where("email = ?", "admin@admin.com").First(&user).Error; err != nil {
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
				RoleID:   id,
			}
			db.Create(&user)
		}
	}
}

func CheckAdmin(c *gin.Context) {
	email, _ := c.Get("userID")
	db, ok := c.MustGet("databaseConn").(*gorm.DB)
	if !ok {
		fmt.Println(ok)
	}
	var user dataModel.User
	var role dataModel.Role
	db.Where("email = ?", email).Find(&user)
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

func helloHandler(c *gin.Context) {
	fmt.Println("hello")
}
