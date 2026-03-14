I'm considering adding to the Temporal CLI an "init" command. I'm thinking something like the following

temporal init https://github.com/temporalio/samples-python/tree/main/hello_standalone_activity
temporal init --python hello_standalone_activity

would both download the sample code, and get a and get a user in position to so run commands from
the README that exists in the dir, typically running a worker, annd then running some starter code.

I think the rule would be that whatever you name, it has to be a diretory containing a README.md,
i.e. a "sample" as Temporal loosely/implicitly defines.

Now, this needs to work for all SDK languages that have samples repos: Go, Java, Python, Dotnet,
Typescript, Ruby, PHP (and in the near future Rust). Ruby and PHP less critical but ideally all
work.  And there's a design question here: do we assume the user is already in a functional language
project or do we initialize the project in the dir we're creating for them?

The purpose here is for a user to quickly be able to get up and running. They run `temporal server
start-dev` (if they're not using cloud), then `temporal init ...`, then run a couple of commands
that are printed to the screen and open the localhost or cloud URL that's printed to the screen to
view the UI.

Remember that we're working in the CLI project that hitherto has known nothing of any language other
than Go. The samples repos are where we define or suggest the toolchain for setting up a project in
each language. But, in the interests of user experience, we are prepared to hard-code some language
logic in the CLI repo if it's inevitable. Ideally I guess the CLI would acquire that logic from some
sort of structured data in the samples repo.

It's essential that we study prior art before designing: how do other technologies address this? I'm
pretty sure I've seen others that have this concept of initializing from a given
example/scenario/samples. Perhaps even some of our competitors, Restate or Inngest?

Finally, we are in the era of AI agents. If all this can be achieved via classical programming then
that would be great. A possibility (not mutually exclusive) is a "skill" that teaches an agent how
to get the given sample up and runnng given the user's on-disk context (maybe they have nothing.
Maybe they're using yarn and not npm. Maybe they're already in a poetry-controlled project. Maybe
they're using something other than gradle. Etc) But I think here I'm asking us what we can achieve
via classical programming. If the conclusion is it's not worth doing anything along these lines with
claassical programming and we should let agents do all this then it's important to know that
conclusion. Always bear in mind that AI agents are going to be ever more sophisticated. But, for the
time being, we also like to provide fast deterministic actions that do not rely on AI.