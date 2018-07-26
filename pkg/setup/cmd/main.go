package main

import (
     "github.com/checkr/flagr/pkg/repo"
     "github.com/checkr/flagr/pkg/setup"
)

// Simple setup program which synchronizes flags from a YAML file to Flagr DB

func main() {
    db := repo.GetDB()
    defer db.Close()

    setup.NewFlagSynchronizer(db).SynchronizeFlags()

}
