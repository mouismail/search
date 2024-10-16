
# Search App

A Go-based application that searches GitHub repositories for specified keywords, utilizing Redis as a caching layer to optimize performance and reduce redundant API calls. The application is containerized using Docker and orchestrated with Docker Compose. Additionally, it includes a GitHub Actions workflow to run the application as a scheduled cron job.

## Table of Contents
- [Features](#features)
- [Prerequisites](#prerequisites)
- [Setup](#setup)
  - [Clone the Repository](#clone-the-repository)
  - [Configure Environment Variables](#configure-environment-variables)
  - [Configure Application](#configure-application)
  - [Running the Application Locally](#running-the-application-locally)
    - [Using Docker Compose](#using-docker-compose)
    - [Without Docker](#without-docker)
- [GitHub Actions Workflow](#github-actions-workflow)
  - [Setting Up GitHub Secrets](#setting-up-github-secrets)
  - [Workflow Configuration](#workflow-configuration)
- [Usage](#usage)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Features
- **GitHub Repository Search**: Fetches repository names and counts based on specified keywords within a GitHub organization.
- **Caching with Redis**: Caches total repository counts, repository names, and keywords to minimize API requests.
- **Containerization**: Dockerfiles for both the application and Redis, managed via Docker Compose.
- **Automated Scheduling**: GitHub Actions workflow to run the application as a scheduled cron job.

## Prerequisites
- Go (version 1.22.3)
- Docker
- Docker Compose
- GitHub Account
- GitHub Token with necessary permissions

## Setup

### Clone the Repository
Clone the repository to your local machine:
```bash
git clone https://github.com/mouismail/search
cd search
```

### Configure Environment Variables
Create a `.env` file in the root directory of the project, or set environment variables directly in your system.

#### Example `.env` file

```bash
GITHUB_TOKEN=your_github_token_here
REDIS_ADDR=redis:6379
ORG_NAME=your-org-name
```

- `GITHUB_TOKEN`: Your GitHub personal access token with the necessary permissions.
- `REDIS_ADDR`: The address where Redis is running (default is localhost:6379).
- `APP_PORT`: The port on which the application will run (default is 8080).

### Configure Application
Edit the `config.yml` file located in the root directory to specify the GitHub organization and the keywords you want to search for within repositories.

#### Example `config.yml`
```yaml
org_name: your_org_name
search_keywords:
  - keyword1
  - keyword2
  - keyword3
```

- `org_name`: The name of the GitHub organization where the search will be performed.
- `keywords`: A list of keywords that the application will use to search for repositories within the specified organization.


### Running the Application Locally

#### Using Docker Compose

To run the application locally using Docker Compose, follow these steps:

##### Build and Start Containers
Run the following command to build the Go application and start both the application and Redis containers:
```bash
docker-compose up --build
```

#### Access the Application

By default, the application will be accessible at `http://localhost:8080`. If you need to change the port, you can do so by editing the `docker-compose.yml` file.

#### Stop Containers

To stop the application and Redis containers, use the following command:
```bash
docker-compose down
```

### Without Docker

#### Install Dependencies

Ensure you have Go installed on your system. The dependencies for the project are managed via `go.mod`. To install them, run the following command:
```bash
go mod tidy
```

#### Run Redis Locally

Make sure Redis is installed and running on your machine. If Redis is not already installed, you can download and install it from [here](https://redis.io/download).

To start Redis, run the following command:
```bash
redis-server
```

#### Set Environment Variables

Before running the application, ensure that the necessary environment variables are set. You can either define them directly in your terminal or use a `.env` file.

For example, to set the environment variables in your terminal:
```bash
export GITHUB_TOKEN=your_github_token_here
export REDIS_ADDR=localhost:6379
export APP_PORT=8080
```

#### Run the Application

Once Redis is running and the environment variables are set, run the application with the following command:
```bash
go run main.go
```

## GitHub Actions Workflow

The project includes a GitHub Actions workflow that runs the application as a scheduled cron job.

### Setting Up GitHub Secrets

#### Navigate to Repository Settings
1. Go to your repository on GitHub and click on **Settings**.

#### Add Secrets
1. Navigate to **Secrets and variables > Actions**.
2. Click on **New repository secret** and add the following:

| Name          | Value                      |
|---------------|----------------------------|
| GITHUB_TOKEN  | your_github_token_here      |

> **Note**: Replace `your_github_token_here` with your actual GitHub token.

### Workflow Configuration

The workflow is defined in `cron.yml`.

#### Adjusting the Cron Schedule
Modify the cron expression in the `cron.yml` file to suit your desired schedule. The current configuration runs the job daily at midnight UTC:
```yaml
schedule:
  - cron: '0 0 * * *'
```

### Committing the Workflow

After configuring the workflow file and setting up secrets, commit the changes to your repository. GitHub Actions will automatically recognize and schedule the workflow based on the defined cron schedule.

## Usage

### Run the Application

- **Locally**: Follow the steps in the [Running the Application Locally](#running-the-application-locally) section.
- **Via Docker Compose**: Use the provided Docker configurations as described earlier.

### Scheduled Runs

The GitHub Actions workflow will execute the application based on the defined schedule, fetching repository data and updating the cache accordingly.

### View Results

Results are printed to the console. To persist or visualize data, consider extending the application to write to a file or integrate it with a monitoring system.

## Troubleshooting

- **Redis Connection Issues**:
  - Ensure Redis is running and accessible at the address specified by `REDIS_ADDR`. If using Docker Compose, verify that the containers are running correctly.

- **GitHub Token Errors**:
  - Verify that the `GITHUB_TOKEN` has the necessary permissions and is correctly set in environment variables or GitHub Secrets.

- **Docker Build Failures**:
  - Ensure Docker and Docker Compose are properly installed and that there are no syntax errors in the `Dockerfile` or `docker-compose.yml`.

- **Application Crashes**:
  - Check the application logs for any errors. Make sure that all dependencies are correctly installed and configured.


  ## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.
