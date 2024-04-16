package main

import (
    "crypto/md5"
    "database/sql"
    "encoding/hex"
    "fmt"
    "net/http"

    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    _ "github.com/go-sql-driver/mysql"
)

type HolidayRequest struct {
    ID       int    `json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Reason   string `json:"reason"`
    FromDate string `json:"from_date"`
    ToDate   string `json:"to_date"`
    Status   string `json:"status"`
}

func main() {
    // Connect to MySQL database
    db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/holiday_request")
    if err != nil {
        fmt.Println("Error connecting to database:", err)
        return
    }
    defer db.Close()

    // Create a new instance of Echo
    e := echo.New()

    // Middleware for CORS
    e.Use(middleware.CORS())

    // Define a route for creating a new user
    e.POST("/createUser", func(c echo.Context) error {
        // Parse JSON request body
        var newUser struct {
            Username string `json:"username"`
            Password string `json:"password"`
            Email    string `json:"email"`
        }
        if err := c.Bind(&newUser); err != nil {
            fmt.Println("Error parsing request body:", err)
            return c.String(http.StatusBadRequest, "Bad request")
        }

        // Hash the password
        hasher := md5.New()
        hasher.Write([]byte(newUser.Password))
        hashedPassword := hex.EncodeToString(hasher.Sum(nil))

        // Check if the username already exists
        var existingUser int
        err := db.QueryRow("SELECT COUNT(*) FROM user WHERE username = ?", newUser.Username).Scan(&existingUser)
        if err != nil {
            fmt.Println("Error checking existing username:", err)
            return c.String(http.StatusInternalServerError, "Internal server error")
        }
        if existingUser > 0 {
            return c.String(http.StatusConflict, "Username already exists")
        }

        // Insert the user into the database with hashed password
        _, err = db.Exec("INSERT INTO user (username, password, email) VALUES (?, ?, ?)", newUser.Username, hashedPassword, newUser.Email)
        if err != nil {
            fmt.Println("Error inserting user:", err)
            return c.String(http.StatusInternalServerError, "Internal server error")
        }

        return c.String(http.StatusOK, "User created successfully")
    })

    // Define a route for user authentication
    e.POST("/user/login", func(c echo.Context) error {
        username := c.FormValue("username")
        password := c.FormValue("password")

        // Hash the password
        hasher := md5.New()
        hasher.Write([]byte(password))
        hashedPassword := hex.EncodeToString(hasher.Sum(nil))

        var storedPassword string
        err := db.QueryRow("SELECT password FROM user WHERE username = ?", username).Scan(&storedPassword)
        if err != nil {
            return c.String(http.StatusInternalServerError, "Internal server error")
        }

        if hashedPassword == storedPassword {
            return c.String(http.StatusOK, "Login successful")
        } else {
            return c.String(http.StatusUnauthorized, "Invalid username or password")
        }
    })

    // Define a route for admin authentication
    e.POST("/admin/login", func(c echo.Context) error {
        username := c.FormValue("username")
        password := c.FormValue("password")

        // Hash the password
        hasher := md5.New()
        hasher.Write([]byte(password))
        hashedPassword := hex.EncodeToString(hasher.Sum(nil))

        var storedPassword string
        err := db.QueryRow("SELECT password FROM admin WHERE username = ?", username).Scan(&storedPassword)
        if err != nil {
            return c.String(http.StatusInternalServerError, "Internal server error")
        }

        if hashedPassword == storedPassword {
            return c.String(http.StatusOK, "Admin login successful")
        } else {
            return c.String(http.StatusUnauthorized, "Invalid username or password")
        }
    })

    // Define a route for creating a new holiday request
    e.POST("/createHolidayRequest", func(c echo.Context) error {
        // Parse JSON request body
        var newRequest struct {
            Name     string `json:"name"`
            Email    string `json:"email"`
            Reason   string `json:"reason"`
            FromDate string `json:"from_date"`
            ToDate   string `json:"to_date"`
        }
        if err := c.Bind(&newRequest); err != nil {
            fmt.Println("Error parsing request body:", err)
            return c.String(http.StatusBadRequest, "Bad request")
        }

        // Insert the holiday request into the database
        _, err := db.Exec("INSERT INTO holiday_requests (name, email, reason, from_date, to_date, status) VALUES (?, ?, ?, ?, ?, ?)", newRequest.Name, newRequest.Email, newRequest.Reason, newRequest.FromDate, newRequest.ToDate, "Pending")
        if err != nil {
            fmt.Println("Error inserting holiday request:", err)
            return c.String(http.StatusInternalServerError, "Internal server error")
        }

        return c.String(http.StatusOK, "Holiday request created successfully")
    })


// Define a route for retrieving pending holiday requests
e.GET("/pending-holiday-requests", func(c echo.Context) error {
    rows, err := db.Query("SELECT id, name, email, reason, from_date, to_date FROM holiday_requests WHERE status = 'Pending'")
    if err != nil {
        fmt.Println("Error querying database for pending holiday requests:", err)
        return c.String(http.StatusInternalServerError, "Internal server error")
    }
    defer rows.Close()

    var pendingRequests []HolidayRequest
    for rows.Next() {
        var request HolidayRequest
        if err := rows.Scan(&request.ID, &request.Name, &request.Email, &request.Reason, &request.FromDate, &request.ToDate); err != nil {
            fmt.Println("Error scanning row:", err)
            continue
        }
        request.Status = "Pending"
        pendingRequests = append(pendingRequests, request)
    }

    if err := rows.Err(); err != nil {
        fmt.Println("Error iterating over rows:", err)
        return c.String(http.StatusInternalServerError, "Internal server error")
    }

    fmt.Println("Pending requests:", pendingRequests)

    return c.JSON(http.StatusOK, pendingRequests)
})
// Define a route for approving a holiday request
e.POST("/approve-request/:id", func(c echo.Context) error {
    id := c.Param("id")

    _, err := db.Exec("UPDATE holiday_requests SET status = 'Approved' WHERE id = ?", id)
    if err != nil {
        fmt.Println("Error updating holiday request status to Approved:", err)
        return c.String(http.StatusInternalServerError, "Internal server error")
    }

    return c.String(http.StatusOK, "Holiday request approved successfully")
})

// Define a route for rejecting a holiday request
e.POST("/reject-request/:id", func(c echo.Context) error {
    id := c.Param("id")

    _, err := db.Exec("UPDATE holiday_requests SET status = 'Rejected' WHERE id = ?", id)
    if err != nil {
        fmt.Println("Error updating holiday request status to Rejected:", err)
        return c.String(http.StatusInternalServerError, "Internal server error")
    }

    return c.String(http.StatusOK, "Holiday request rejected successfully")
})



    // Start the server
    e.Logger.Fatal(e.Start(":8080"))
}