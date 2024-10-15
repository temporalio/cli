// Package commandsgen is built to read the YAML format described in
// temporalcli/commandsgen/commands.yml and generate code from it.
package commandsgen

import (
	_ "embed"
	"sort"
	"strings"
)

func EnrichCommands(m Commands) (Commands, error) {
	commandLookup := make(map[string]*Command)

	for i, command := range m.CommandList {
		m.CommandList[i].Index = i
		commandLookup[command.FullName] = &m.CommandList[i]
	}

	var rootCommand *Command

	//populate parent and basic meta
	for i, c := range m.CommandList {
		commandLength := len(strings.Split(c.FullName, " "))
		if commandLength == 1 {
			rootCommand = &m.CommandList[i]
			continue
		}
		parentName := strings.Join(strings.Split(c.FullName, " ")[:commandLength-1], " ")
		parent, ok := commandLookup[parentName]
		if ok {
			m.CommandList[i].Parent = &m.CommandList[parent.Index]
			m.CommandList[i].Depth = len(strings.Split(c.FullName, " ")) - 1
			m.CommandList[i].FileName = strings.Split(c.FullName, " ")[1]
			m.CommandList[i].LeafName = strings.Join(strings.Split(c.FullName, " ")[m.CommandList[i].Depth:], "")
		}
	}

	//populate children and base command
	for _, c := range m.CommandList {
		if c.Parent == nil {
			continue
		}

		//fmt.Printf("add child: %s\n", m.CommandList[c.Index].FullName)
		m.CommandList[c.Parent.Index].Children = append(m.CommandList[c.Parent.Index].Children, &m.CommandList[c.Index])

		base := &c
		for base.Depth > 1 {
			base = base.Parent
		}
		m.CommandList[c.Index].Base = &m.CommandList[base.Index]
	}

	setMaxChildDepthVisitor(*rootCommand, &m)

	for i, c := range m.CommandList {
		if c.Parent == nil {
			continue
		}

		subCommandStartDepth := 1
		if c.Base.MaxChildDepth > 2 {
			subCommandStartDepth = 2
		}

		subCommandName := ""
		if c.Depth >= subCommandStartDepth {
			subCommandName = strings.Join(strings.Split(c.FullName, " ")[subCommandStartDepth:], " ")
		}

		if len(subCommandName) == 0 && c.Depth == 1 {
			// for operator base command to show up in tags, keywords, etc.
			subCommandName = c.LeafName
		}

		m.CommandList[i].SubCommandName = subCommandName
	}

	// sorted ascending by full name of command (activity complete, batch list, etc)
	sortChildrenVisitor(rootCommand)

	// pull flat list in same order as sorted children
	m.CommandList = make([]Command, 0)
	collectCommandVisitor(*rootCommand, &m)

	// option usages
	optionUsages := getAllOptionUsages(m)
	optionUsagesByOptionDescription := getOptionUsagesByOptionDescription(optionUsages)
	m.Usages = Usages{
		OptionUsages:                    optionUsages,
		OptionUsagesByOptionDescription: optionUsagesByOptionDescription,
	}

	return m, nil
}

func collectCommandVisitor(c Command, m *Commands) {

	m.CommandList = append(m.CommandList, c)

	for _, child := range c.Children {
		collectCommandVisitor(*child, m)
	}
}

func sortChildrenVisitor(c *Command) {
	sort.Slice(c.Children, func(i, j int) bool {
		//option to put nested commands at end of the list
		/*
			if c.Children[i].MaxChildDepth != c.Children[j].MaxChildDepth {
				return c.Children[i].MaxChildDepth < c.Children[j].MaxChildDepth
			}
		*/

		return c.Children[i].FullName < c.Children[j].FullName
	})
	for _, command := range c.Children {
		sortChildrenVisitor(command)
	}
}

func setMaxChildDepthVisitor(c Command, commands *Commands) int {
	maxChildDepth := 0
	children := commands.CommandList[c.Index].Children
	if len(children) > 0 {
		for _, child := range children {
			depth := setMaxChildDepthVisitor(*child, commands)
			if depth > maxChildDepth {
				maxChildDepth = depth
			}
		}
	}

	commands.CommandList[c.Index].MaxChildDepth = maxChildDepth
	return maxChildDepth + 1
}

func getAllOptionUsages(commands Commands) []OptionUsages {
	// map[optionName]map[usageSite]OptionUsageSite
	var optionUsageSitesMap = make(map[string]map[string]OptionUsageSite)

	// option sets
	for i, optionSet := range commands.OptionSets {
		usage := optionSet.Description
		if len(usage) == 0 {
			usage = optionSet.Name
		}

		for j, option := range optionSet.Options {
			_, found := optionUsageSitesMap[option.Name]
			if !found {
				optionUsageSitesMap[option.Name] = make(map[string]OptionUsageSite)
			}
			optionUsageSitesMap[option.Name][optionSet.Name] = OptionUsageSite{
				Option:               commands.OptionSets[i].Options[j],
				UsageSiteDescription: usage,
				UsageSiteType:        UsageTypeOptionSet,
			}
		}
	}

	//command options
	for i, cmd := range commands.CommandList {
		usage := cmd.FullName
		if len(usage) == 0 {
			usage = cmd.FullName
		}

		for j, option := range cmd.Options {
			_, found := optionUsageSitesMap[option.Name]
			if !found {
				optionUsageSitesMap[option.Name] = make(map[string]OptionUsageSite)
			}
			optionUsageSitesMap[option.Name][cmd.FullName] = OptionUsageSite{
				Option:               commands.CommandList[i].Options[j],
				UsageSiteDescription: usage,
				UsageSiteType:        UsageTypeOptionSet,
			}
		}
	}

	// all options
	var allOptionUsages = make([]OptionUsages, 0)

	for optionName, usages := range optionUsageSitesMap {
		option := OptionUsages{
			OptionName: optionName,
			UsageSites: make([]OptionUsageSite, 0),
		}
		for _, usage := range usages {
			option.UsageSites = append(option.UsageSites, usage)
		}
		allOptionUsages = append(allOptionUsages, option)
	}

	sort.Slice(allOptionUsages, func(i, j int) bool {
		return allOptionUsages[i].OptionName < allOptionUsages[j].OptionName
	})

	for u := range allOptionUsages {
		sort.Slice(allOptionUsages[u].UsageSites, func(i, j int) bool {
			return allOptionUsages[u].UsageSites[i].UsageSiteDescription < allOptionUsages[u].UsageSites[j].UsageSiteDescription
		})
	}

	return allOptionUsages
}

func getOptionUsagesByOptionDescription(allOptionUsages []OptionUsages) []OptionUsagesByOptionDescription {
	out := make([]OptionUsagesByOptionDescription, len(allOptionUsages))

	for i, optionUsages := range allOptionUsages {
		out[i].OptionName = optionUsages.OptionName

		if len(optionUsages.UsageSites) == 1 {
			usage := allOptionUsages[i].UsageSites[0]
			out[i].Usages = make([]OptionUsageByOptionDescription, 1)
			out[i].Usages[0].OptionDescription = usage.Option.Description
			out[i].Usages[0].UsageSites = []OptionUsageSite{usage}

			continue
		}

		// map[optionDescription]OptionUsageByOptionDescription
		optionUsageByOptionDescriptionMap := make(map[string]OptionUsageByOptionDescription)

		// collate on option description in each usage site
		for j, usage := range optionUsages.UsageSites {
			_, found := optionUsageByOptionDescriptionMap[usage.Option.Description]
			if !found {
				optionUsageByOptionDescriptionMap[usage.Option.Description] = OptionUsageByOptionDescription{
					OptionDescription: usage.Option.Description,
					UsageSites:        make([]OptionUsageSite, 0),
				}
			}
			u := optionUsageByOptionDescriptionMap[usage.Option.Description]
			u.UsageSites = append(u.UsageSites, allOptionUsages[i].UsageSites[j])

			// put all distinct option descriptions withing the option usages
			optionUsageByOptionDescriptionMap[u.OptionDescription] = u
		}

		out[i].Usages = make([]OptionUsageByOptionDescription, len(optionUsageByOptionDescriptionMap))
		j := 0
		for _, v := range optionUsageByOptionDescriptionMap {
			out[i].Usages[j] = v
			j++
		}
	}

	return out
}
