Small to Medium Project (Flat / Standard Layout)

my-api/
├── config/             # จัดการเรื่อง Environment variables & App Config
│   └── config.go
├── db/                 # จัดการ Database Connection
│   └── database.go
├── models/             # GORM Models / Structs Data Type
│   └── album.go
├── handlers/           # Gin Route Handlers / Controllers
│   └── album_handler.go
├── middleware/
│   └── logger.go       # Middleware logger
├── router/
│   └──router.go        # รวมการตั้งค่า Gin Routes
├── go.mod
├── go.sum 
└── main.go             # Entry point หลัก (อ่าน config, connection DB, สั่ง run server)


command execute build 
-install Go De-Bugger Attach to Process 
-go build -gcflags="all=-N -l" -o manga_app.exe
-.\manga_app.exe


