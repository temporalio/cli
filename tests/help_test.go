package tests

func (s *e2eSuite) TestHelp_PrintsCommands() {
	s.T().Parallel()

	testserver, app, writer := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	err := app.Run([]string{"", "--help"})
	s.NoError(err)

	tests := []struct {
		want string
	}{
		{"workflow"},
		{"Start, list, and operate on Workflows"},
		{"schedule"},
		{"Create and edit Schedules"},
	}

	for _, tt := range tests {
		s.Contains(writer.GetContent(), tt.want)
	}
}

func (s *e2eSuite) TestHelp_PrintsSubcommands() {
	s.T().Parallel()

	testserver, app, writer := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	err := app.Run([]string{"", "workflow", "--help"})
	s.NoError(err)

	tests := []struct {
		want string
	}{
		{"start"},
		{"Starts a new Workflow Execution"},
		{"list"},
		{"List Workflow Executions based on a Query"},
	}

	for _, tt := range tests {
		s.Contains(writer.GetContent(), tt.want)
	}
}

func (s *e2eSuite) TestHelp_PrintsFlags() {
	s.T().Parallel()

	testserver, app, writer := s.setUpTestEnvironment()
	defer func() {
		_ = testserver.Stop()
	}()

	err := app.Run([]string{"", "workflow", "list", "--help"})
	s.NoError(err)

	tests := []struct {
		want string
	}{
		{"Command Options:"},
		{"--query"},
		{"Filter results using an SQL-like query"},
		{"Display Options:"},
		{"--limit"},
		{"Number of items to print"},
		{"Shared Options:"},
		{"--address"},
		{"The host and port"},
	}

	for _, tt := range tests {
		s.Contains(writer.GetContent(), tt.want)
	}
}
