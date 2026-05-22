import asyncio
from config import TARGET_TOPICS
from classifier import classify_file
from messaging import NatsMessagingAdapter

def core_batch_processing_workflow(files_list):

    results = []
    for f in files_list:
        res = classify_file(
            filepath=f["filepath"], 
            topics=TARGET_TOPICS, 
            mime_type=f["mime_type"]
        )
        
        results.append({
            "filepath": res["filepath"],
            "topTopic": res["top_topic"],
            "confidence": res["confidence"]
        })
    return results

async def main():
    # Initialize the NATS network infrastructure adapter
    adapter = NatsMessagingAdapter(processing_callback=core_batch_processing_workflow)
    
    # Establish connection bounds and subscribe using the shared worker queue pool.
    await adapter.start(subject="llm.classify.batch", queue_group="llm_processing_pool")
    
    print("Python classification engine ready. Waiting for tasks from Go...")
    await adapter.keep_alive()

if __name__ == '__main__':
    asyncio.run(main())