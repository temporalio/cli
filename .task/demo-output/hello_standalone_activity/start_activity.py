import asyncio
from datetime import timedelta

from temporalio.client import Client
from temporalio.envconfig import ClientConfig

from hello_standalone_activity.my_activity import ComposeGreetingInput, compose_greeting


async def my_application():
    connect_config = ClientConfig.load_client_connect_config()
    connect_config.setdefault("target_host", "localhost:7233")
    client = await Client.connect(**connect_config)

    # Start the activity without waiting for the result
    activity_handle = await client.start_activity(
        compose_greeting,
        args=[ComposeGreetingInput("Hello", "World")],
        id="my-standalone-activity-id",
        task_queue="my-standalone-activity-task-queue",
        start_to_close_timeout=timedelta(seconds=10),
    )
    print(f"Started activity: {activity_handle.id}")

    # Wait for the result
    activity_result = await activity_handle.result()
    print(f"Activity result: {activity_result}")


if __name__ == "__main__":
    asyncio.run(my_application())
