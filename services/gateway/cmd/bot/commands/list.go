package commands

// List lists all bots
func List(args []string) {
	// Initialize database and service
	db, err := initDB()
	if err != nil {
		fatal("Failed to initialize database: %v", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	botService, err := initBotService(db)
	if err != nil {
		fatal("Failed to initialize bot service: %v", err)
	}

	// List bots
	ctx, cancel := getContext()
	defer cancel()

	bots, err := botService.ListBots(ctx, 0, 1000)
	if err != nil {
		fatal("Failed to list bots: %v", err)
	}

	printBotTable(bots)
}
