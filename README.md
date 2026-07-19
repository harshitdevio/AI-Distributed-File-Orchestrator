# Ark1

Ark1 is a local AI-powered file organizer.

It scans a directory, classifies files using a zero-shot LLM (facebook/bart-large-mnli), and orchestrates them into suitable directories as per the data provided by the LLM. The project is split into independent, decoupled services that communicate over NATS, allowing the orchestration and the inference layers to operate & evolve separately.

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
---
> **Note**
>
> This project was built primarily to demonstrate system design and software architecture. As a result, it intentionally prioritizes backend architecture, service boundaries, communication patterns, and engineering trade-offs over user interface or user experience.
>
> The primary objective of Ark1 is to demonstrate architectural thinking and engineering decisions instead of building a polished end-user application.
>
> I am currently working on an alternative implementation of the same application using a different architecture with a different set of constraints and trade-offs to compare design approaches. Once completed, it will be linked here.

## License

This project is licensed under the Apache License 2.0. See the LICENSE file for details.
