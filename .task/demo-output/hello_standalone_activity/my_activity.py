from dataclasses import dataclass

from temporalio import activity


@dataclass
class ComposeGreetingInput:
    greeting: str
    name: str


@activity.defn
def compose_greeting(input: ComposeGreetingInput) -> str:
    activity.logger.info("Running activity with parameter %s" % input)
    return f"{input.greeting}, {input.name}!"
