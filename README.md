# Veil

Veil is an open-source program written in Golang designed to efficiently scrape, process, and manage class and college enrollment data. In addition to offering a seamless way to search and export class data, it also supports class enrollment.

## Table of Contents

- [Key Features](#key-features)
- [Prerequisites](#prerequisites)
- [Configuration](#configuration)
- [Compilation](#compilation)
- [Usage](#usage)
- [Notification Example](#notification-example)

## Key Features

1. **Class Search & Export**: Ability to search for classes and export the results in CSV format.
2. **Unofficial Transcript**: Retrieve and export your previously enrolled courses in CSV format.
3. **Enrollment**: Enroll in courses.

## Prerequisites

- **Golang**: You need a version >=1.18.3 of [Go](https://go.dev/doc/install) installed.

## Configuration 

For the tool to function correctly, a `.env` file is required with the following parameters:

### .env Parameters

| Parameter      | Description                                         | Example Values          |
|----------------|-----------------------------------------------------|-------------------------|
| CAMPUSID       | Your login ID                                       |                         |
| PASSWORD       | Your login password                                 |                         |
| MODE           | Operation mode (SIGNUP, SEARCH, or EXPORT)          | `MODE=SIGNUP`           |
| SUBJECT        | Course subject to search for                        | `SUBJECT=PHYS`          |
| YEAR           | Target academic year                                | `YEAR=2024`              |
| QUARTER        | Target academic quarter                             | `QUARTER=WINTER`        |
| CAMPUS         | Campus code (either DA or FH)                       | `CAMPUS=DA`             |
| CRNTOADD       | Course Reference Numbers for enrollment (SIGNUP mode) seperated by comma | `CRNTOADD=00000,00001`        |
| RETRY_AMOUNT   | Max number of retry attempts                        | `RETRY_AMOUNT=2`        |
| RETRY_DURATION | Duration to wait between retries (in seconds)       | `RETRY_DURATION=2`      |
| DISCORD_WEBHOOK| Discord notification webhook                        |                         |

**Note**: Ensure you keep the `.env` file secure, as it contains sensitive login credentials which are not encrypted.

## Compilation

To compile the program, run compile.sh

```bash
./compile.sh
```

## Usage

Based on the `MODE` set in the `.env` file:

- **SIGNUP**: Enroll in the class with the specified CRN.
- **SEARCH**: Search for classes based on the given term, section, and subject.
- **EXPORT**: Export details of all previously enrolled courses.

## Notification Example

![Notification](https://cdn.discordapp.com/attachments/1022240002408730644/1168028448921497620/image.png)
