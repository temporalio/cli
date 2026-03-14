import asyncio

from temporalio.client import Client
from temporalio.envconfig import ClientConfig


async def my_application():
    connect_config = ClientConfig.load_client_connect_config()
    connect_config.setdefault("target_host", "localhost:7233")
    client = await Client.connect(**connect_config)

    activities = client.list_activities(
        query="TaskQueue = 'my-standalone-activity-task-queue'",
    )

    async for info in activities:
        print(
            f"ActivityID: {info.activity_id}, Type: {info.activity_type}, Status: {info.status}"
        )


if __name__ == "__main__":
    asyncio.run(my_application())
