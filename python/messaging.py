import asyncio
import json
import os
import sys
from nats.aio.client import Client as NATS

class NatsMessagingAdapter:
    def __init__(self, processing_callback):
        """
        Accepts a clean execution callback function from the core app logic domain.
        The callback should expect a raw parsed Python dictionary/list and return one.
        """
        self.nc = NATS()
        self.processing_callback = processing_callback
        self.nats_url = os.environ.get("NATS_URL", "nats://localhost:4222")

    async def _message_handler(self, msg):
        try:
            # Protocol Layer: extract & translate inbound wire serialization format
            inbound_data = json.loads(msg.data.decode())
            
            # Application Logic Domain Layer: completely decoupled execution
            outbound_domain_data = self.processing_callback(inbound_data)
            
            # Protocol Layer: translate & package outbound serialization representation
            response_bytes = json.dumps(outbound_domain_data).encode()
            await msg.respond(response_bytes)
            
        except Exception as err:
            print(f"[Messaging Adapter Error] Processing breakdown: {err}", file=sys.stderr)
            await msg.respond(b"[]") # Safeguard boundary acknowledgment

    async def start(self, subject: str, queue_group: str):
        try:
            await self.nc.connect(self.nats_url)
            print(f"[Network Fabric] Connection verified on {self.nats_url}")
            
            await self.nc.subscribe(
                subject=subject,
                queue=queue_group,
                cb=self._message_handler
            )
            print(f"[Network Fabric] Subscribed to '{subject}' using queue group '{queue_group}'")
        except Exception as conn_err:
            print(f"[Network Fabric Critical Failure] Cannot attach to broker: {conn_err}", file=sys.stderr)
            sys.exit(1)

    async def keep_alive(self):
        while True:
            await asyncio.sleep(1)