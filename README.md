# Gent News RSS Feed Generator

This service automatically generates and serves an RSS feed from the Stad Gent news API. It fetches the latest news articles and converts them into a standardized RSS format, making it easy to follow Gent's news updates through any RSS reader.

## Features

- Automatically fetches news from Stad Gent's API
- Converts JSON data to RSS 2.0 format
- Updates feed every hour
- Serves RSS feed via HTTP
- Dockerized for easy deployment

## API Endpoint

The RSS feed is available at:
```
http://localhost:8080/feed
```

## Docker Instructions

### Building the Image

To build the Docker image, run:
```bash
docker build -t gent-news-rss .
```

### Running the Container

To run the container:
```bash
docker run -p 8080:8080 gent-news-rss
```

This will:
- Start the RSS feed generator
- Make the feed available at http://localhost:8080/feed
- Automatically update the feed every hour

### Running in Background

To run the container in the background:
```bash
docker run -d -p 8080:8080 gent-news-rss
```

### Viewing Logs

To view the container logs:
```bash
docker logs gent-news-rss
```

To follow the logs in real-time:
```bash
docker logs -f gent-news-rss
```

## Technical Details

- Built with Go
- Uses the Stad Gent API for news data
- Implements RSS 2.0 specification
- Runs on port 8080 by default
- Updates feed every hour
- Uses Alpine Linux for minimal container size

## RSS Feed Information

The generated RSS feed includes:
- Title
- Link
- Description
- Publication Date
- WebMaster contact
- Individual news items with:
  - Title
  - Link
  - Description
  - Publication Date
  - GUID

## License

This project is open source and available under the MIT License. 