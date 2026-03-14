import asyncio
from concurrent.futures import ThreadPoolExecutor

from temporalio.client import Client
from temporalio.envconfig import ClientConfig
from temporalio.worker import Worker

from hello_standalone_activity.my_activity import compose_greeting


async def main():
    connect_config = ClientConfig.load_client_connect_config()
    connect_config.setdefault("target_host", "localhost:7233")
    client = await Client.connect(**connect_config)
    worker = Worker(
        client,
        task_queue="my-standalone-activity-task-queue",
        activities=[compose_greeting],
        activity_executor=ThreadPoolExecutor(5),
    )
    print("worker running...", end="", flush=True)
    await worker.run()


if __name__ == "__main__":
    asyncio.run(main())
