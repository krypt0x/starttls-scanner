package db_test

import (
    "fmt"
    "log"
    "os"
    "testing"
    "time"

    "github.com/sydneyli/starttls-scanner/db"
)

// Global database object for tests.
var database *db.SqlDatabase

// Connects to local test db.
func initTestDb() *db.SqlDatabase {
    os.Setenv("PRIV_KEY", "./certs/key.pem")
    os.Setenv("PUBLIC_KEY", "./certs/cert.pem")
    cfg, err := db.LoadEnvironmentVariables()
    cfg.Db_name = fmt.Sprintf("%s_dev", cfg.Db_name)
    if err != nil {
        log.Fatal(err)
    }
    database, err := db.InitSqlDatabase(cfg)
    if err != nil {
        log.Fatal(err)
    }
    return database
}

func TestMain(m *testing.M) {
    database = initTestDb()
    code := m.Run()
    err := database.ClearTables()
    if err != nil {
        log.Fatal(err)
    }
    os.Exit(code)
}

///////////////////////////
// ***** API tests ***** //
///////////////////////////

// TODO (sydli)

////////////////////////////////
// ***** Database tests ***** //
////////////////////////////////

func TestPutScan(t *testing.T) {
    database.ClearTables()
    dummyScan := db.ScanData{
        Domain: "dummy.com",
        Data: "{}",
        Timestamp: time.Now(),
    }
    err := database.PutScan(dummyScan)
    if err != nil {
        t.Errorf("PutScan failed: %v\n", err)
    }
}

func TestGetLatestScan(t *testing.T) {
    database.ClearTables()
    // Add two dummy objects
    earlyScan := db.ScanData{
        Domain: "dummy.com",
        Data: "test_before",
        Timestamp: time.Now(),
    }
    laterScan := db.ScanData{
        Domain: "dummy.com",
        Data: "test_after",
        Timestamp: time.Now().Add(time.Duration(time.Hour)),
    }
    err := database.PutScan(laterScan)
    if err != nil {
        t.Errorf("PutScan failed: %v\n", err)
    }
    err = database.PutScan(earlyScan)
    if err != nil {
        t.Errorf("PutScan failed: %v\n", err)
    }
    scan, err := database.GetLatestScan("dummy.com")
    if err != nil {
        t.Errorf("GetLatestScan failed: %v\n", err)
    }
    if scan.Data != "test_after" {
        t.Errorf("Expected GetLatestScan to retrieve most recent scanData: %v", scan)
    }
}

func TestGetAllScans(t *testing.T) {
    database.ClearTables()
    data, err := database.GetAllScans("dummy.com")
    if err != nil {
        t.Errorf("GetAllScans failed: %v\n", err)
    }
    // Retrieving scans for domain that's never been scanned before
    if len(data) != 0 {
        t.Errorf("Expected GetAllScans to return []")
    }
    // Add two dummy objects
    dummyScan := db.ScanData{
        Domain: "dummy.com",
        Data: "test1",
        Timestamp: time.Now(),
    }
    err = database.PutScan(dummyScan)
    if err != nil {
        t.Errorf("PutScan failed: %v\n", err)
    }
    dummyScan.Data = "test2"
    err = database.PutScan(dummyScan)
    if err != nil {
        t.Errorf("PutScan failed: %v\n", err)
    }
    data, err = database.GetAllScans("dummy.com")
    // Retrieving scans for domain that's been scanned once
    if err != nil {
        t.Errorf("GetAllScans failed: %v\n", err)
    }
    if len(data) != 2 {
        t.Errorf("Expected GetAllScans to return two items, returned %d\n", len(data))
    }
    if data[0].Data != "test1" || data[1].Data != "test2" {
        t.Errorf("Expected Data of scan objects to include both test1 and test2")
    }
}

func TestPutGetDomain(t *testing.T) {
    database.ClearTables()
    data := db.DomainData {
        Name: "testing.com",
        Email: "admin@testing.com",
    }
    err := database.PutDomain(data)
    if err != nil {
        t.Errorf("PutDomain failed: %v\n", err)
    }
    retrieved_data, err := database.GetDomain(data.Name)
    if err != nil {
        t.Errorf("GetDomain(%s) failed: %v\n", data.Name, err)
    }
    if retrieved_data.Name != data.Name {
        t.Errorf("Somehow, GetDomain retrieved the wrong object?")
    }
    if retrieved_data.State != db.StateUnvalidated {
        t.Errorf("Default state should be 'Unvalidated'")
    }
}

func TestUpsertDomain(t *testing.T) {
    database.ClearTables()
    data := db.DomainData {
        Name: "testing.com",
        Email: "admin@testing.com",
    }
    database.PutDomain(data)
    err := database.PutDomain(db.DomainData { Name: "testing.com", State: db.StateQueued })
    if err != nil {
        t.Errorf("PutDomain(%s) failed: %v\n", data.Name, err)
    }
    retrieved_data, err := database.GetDomain(data.Name)
    if retrieved_data.State != db.StateQueued {
        t.Errorf("Expected state to be 'Queued', was %v\n", retrieved_data)
    }
}

func TestPutUseToken(t *testing.T) {
    database.ClearTables()
    data, err := database.PutToken("testing.com")
    if err != nil {
        t.Errorf("PutToken failed: %v\n", err)
    }
    domain, err := database.UseToken(data.Token)
    if err != nil {
        t.Errorf("UseToken failed: %v\n", err)
    }
    if domain != data.Domain {
        t.Errorf("UseToken used token for %s instead of %s\n", domain, data.Domain)
    }
}

