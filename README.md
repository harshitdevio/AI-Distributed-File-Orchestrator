# Ark1

Ark1 is a local AI-powered file organizer.

It scans a directory, classifies files using a zero-shot LLM (facebook/bart-large-mnli), and orchestrates them into suitable directories as per the data provided by the LLM. The project is split into independent decoupled services that communicate over NATS, allowing the orchestration layer and the inference layer to operate & evolve separately.

## Overview

- Orchestration service written in Go
- LLM inference service written in Python
- NATS used for inter-service communication
- Zero-shot document classification (no task-specific model training)

## High-Level Design 

<img width="1408" height="768" alt="wmremove-transformed (1)" src="https://github.com/user-attachments/assets/175c4a0c-fa6b-4bb5-acc8-4e7873013ff3" />

## Documentation

The README intentionally stays brief.

For the detailed system architecture, execution flow, design decisions, trade-offs, and implementation details, see the project documentation:

**→ [Architecture Documentation](https://harshitdevio.gitbook.io/ark1/page/ark1)**

## License

This project is licensed under the Apache License 2.0. See the LICENSE file for details.
