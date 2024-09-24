package commandsgen

// TODO: figure out how to generate each file:
//
//	activity, batch, cmd-options, env, index,
//	operator, schedule, server, task-queue, workflow
// cmd-options -
//     - tags: not sure where this comes from

// index doesn't need to be generated

type DocsFile struct {
	FileName string
	Data     []byte
}

func GenerateDocsFiles(commands Commands) ([]DocsFile, error) {
	return []DocsFile{}, nil
}
