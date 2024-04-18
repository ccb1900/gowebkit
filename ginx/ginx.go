package ginx

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/ccb1900/gocommon/config"
	"github.com/ccb1900/gocommon/logger"
	"github.com/ccb1900/gocommon/ulidx"
	"github.com/ccb1900/gowebkit"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

type App struct {
	AuthRoutes        func(*gin.RouterGroup)
	Routes            func(*gin.Engine)
	AuthJwtMiddleware func(*gin.Engine, *jwt.GinJWTMiddleware)
	assets            embed.FS
	ui                bool
	uipath            string
	port              int
	name              string
}

var Auth *jwt.GinJWTMiddleware

func New(name string, port int) *App {
	return &App{
		port: port,
		name: name,
		AuthRoutes: func(gn *gin.RouterGroup) {
			logger.Default().Warn("auth routes not set")
		},
		Routes: func(gn *gin.Engine) {
			logger.Default().Warn("routes not set")
		},
		// AuthJwtMiddleware: func(e *gin.Engine, jwt *jwt.GinJWTMiddleware) {
		// 	logger.Default().Warn("auth jwt middleware not set")
		// },
	}
}

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

type User struct {
	UserName  string
	FirstName string
	LastName  string
}

type IUser interface {
	UserName() string
	Id() string
}

var ErrDir = errors.New("path is dir")

func (a *App) SetUI(assets embed.FS, path string) {
	a.assets = assets
	a.uipath = path
	a.ui = true
}

func (a *App) tryRead(fs embed.FS, prefix, requestedPath string, w http.ResponseWriter) error {
	f, err := fs.Open(path.Join(prefix, requestedPath))
	if err != nil {
		logger.Default().Info("try read", "path", path.Join(prefix, requestedPath), "err", err)
		return err
	}
	defer f.Close()

	stat, _ := f.Stat()
	if stat.IsDir() {
		return ErrDir
	}

	contentType := mime.TypeByExtension(filepath.Ext(requestedPath))
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", contentType)

	_, err = io.Copy(w, f)
	if err != nil {
		return err
	}
	return nil
}

func (app *App) Run(ctx context.Context) error {
	if !config.Default().GetBool("debug") {
		gin.SetMode(gin.ReleaseMode)
	}

	gn := gin.New()
	gn.Use(gin.Logger())
	gn.Use(gin.Recovery())
	gn.Use(cors.Default())
	gn.Use(requestid.New())
	secret := config.Default().GetString("jwt.secret")
	identityKey := "id"
	var err error
	Auth, err = jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "dftc",
		Key:         []byte(secret),
		Timeout:     14 * 24 * time.Hour,
		MaxRefresh:  0,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			fmt.Println("payload", data)
			if v, ok := data.(IUser); ok {
				fmt.Println("payload2", data)
				return jwt.MapClaims{
					identityKey: v.Id(),
				}
			}
			return jwt.MapClaims{}
		},
		// IdentityHandler: func(c *gin.Context) interface{} {
		// 	claims := jwt.ExtractClaims(c)
		// 	return &User{
		// 		UserName: claims[identityKey].(string),
		// 	}
		// },
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals login
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			userID := loginVals.Username
			password := loginVals.Password

			if (userID == "admin" && password == "admin") || (userID == "test" && password == "test") {
				return &User{
					UserName:  userID,
					LastName:  "Bo-Yi",
					FirstName: "Wu",
				}, nil
			}

			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			// if v, ok := data.(*User); ok && v.UserName == "admin" {
			// 	return true
			// }

			// return false
			return true
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		// - "param:<name>"
		TokenLookup: "header: Authorization",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})
	if err != nil {
		log.Fatal("JWT Error:" + err.Error())
	}

	if config.Default().GetBool("upload.enable") {
		gn.POST("upload/single/local", func(ctx *gin.Context) {
			file, err := ctx.FormFile("file")
			if err != nil {
				logger.Default().Error("upload file error", "err", err)
				ctx.JSON(400, gowebkit.ResultError("参数错误"))
				return
			}

			fileName := fmt.Sprintf("%s%s", ulidx.Get(), path.Ext(file.Filename))
			pathName := fmt.Sprintf("%s/%s", config.Default().GetString("upload.path"), fileName)
			if err := ctx.SaveUploadedFile(file, pathName); err != nil {
				logger.Default().Error("save file error", "err", err)
				ctx.JSON(400, gowebkit.ResultError("参数错误"))
				return
			}

			ctx.JSON(http.StatusOK, gowebkit.ResultOkData(fileName))
		})
		gn.POST("upload/many/local", func(ctx *gin.Context) {
			form, err := ctx.MultipartForm()
			if err != nil {
				logger.Default().Error("upload file error", "err", err)
				ctx.JSON(400, gowebkit.ResultError("参数错误"))
				return
			}
			files := form.File["files[]"]

			var results []string

			for _, file := range files {
				// Upload the file to specific dst.
				fileName := fmt.Sprintf("%s%s", ulidx.Get(), path.Ext(file.Filename))
				pathName := fmt.Sprintf("%s/%s", config.Default().GetString("upload.path"), fileName)
				if err := ctx.SaveUploadedFile(file, pathName); err != nil {
					logger.Default().Error("save file error", "err", err)
					ctx.JSON(400, gowebkit.ResultError("参数错误"))
					return
				}

				results = append(results, fileName)
			}
			ctx.JSON(http.StatusOK, gowebkit.ResultOkData(results))
		})
	}

	app.Routes(gn)
	if app.AuthJwtMiddleware != nil {
		app.AuthJwtMiddleware(gn, Auth)
		gn.POST("/login", Auth.LoginHandler)
		authorized := gn.Group("/")
		authorized.Use(Auth.MiddlewareFunc())
		{
			app.AuthRoutes(authorized)
		}
	} else {
		logger.Default().Warn("auth jwt middleware not set")
	}
	gn.NoMethod(func(ctx *gin.Context) {
		ctx.JSON(405, gin.H{
			"code":    405,
			"message": "method not supported",
		})
	})
	gn.NoRoute(func(ctx *gin.Context) {
		log.Println("404")
		if !app.ui {
			ctx.JSON(404, gin.H{
				"code":    404,
				"message": "not found",
			})
			return
		}
		if err := app.tryRead(app.assets, app.uipath, ctx.Request.URL.Path, ctx.Writer); err == nil {
			return
		} else {
			logger.Default().Info("加载失败", "path", ctx.Request.URL.Path, "err", err)
		}

		if err := app.tryRead(app.assets, app.uipath, "index.html", ctx.Writer); err != nil {
			return
		}
	})

	s := http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Default().GetString("host"), app.port),
		Handler: gn,
	}
	go func(ctx context.Context) {
		defer log.Println("boot server exit", "name", app.name)
		<-ctx.Done()
		if err := s.Shutdown(ctx); err != nil {
			logger.Default().Error("boot server shutdown fail", "err", err)
		}
	}(ctx)

	log.Println("boot server", "name", app.name, "port", app.port)
	if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Default().Error("boot server fail", "err", err)
		return err
	}

	return nil
}
