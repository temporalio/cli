import asyncio

from temporalio.client import Client
from temporalio.envconfig import ClientConfig


async def my_application():
    connect_config = ClientConfig.load_client_connect_config()
    connect_config.setdefault("target_host", "localhost:7233")
    client = await Client.connect(**connect_config)

    resp = await client.count_activities(
        query="TaskQueue = 'my-standalone-activity-task-queue'",
    )

    print("Total activities:", resp.count)

    for group in resp.groups:
        print(f"Group {group.group_values}: {group.count}")


if __name__ == "__main__":
    asyncio.run(my_application())
