<!-- Improved compatibility of back to top link: See: https://github.com/othneildrew/Best-README-Template/pull/73 -->

<a name="readme-top"></a>

<!--
*** Thanks for checking out the Best-README-Template. If you have a suggestion
*** that would make this better, please fork the repo and create a pull request
*** or simply open an issue with the tag "enhancement".
*** Don't forget to give the project a star!
*** Thanks again! Now go create something AMAZING! :D
-->

<!-- PROJECT LOGO -->
<br />
<div align="center">
  <h3 align="center">Film Sync</h3>

  <p align="center">
    A Go application to help manage my film photos
    <br />
    <a href="https://github.com/ftrbnd/film-sync/issues">Report Bug</a>
    ·
    <a href="https://github.com/ftrbnd/film-sync/issues">Request Feature</a>
  </p>
</div>

<!-- TABLE OF CONTENTS -->
<details>
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
        <li><a href="#configuration">Configuration</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#contact">Contact</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->

## About The Project

Welcome to my first Go project!

When I get my film photos developed by my local photo studio, they email me the scans via a WeTransfer link. I have been manually opening the link, extracting the photos from the zip file, and uploading the .TIF files to my Google Drive. After all this, I still don't have shareable photos since most apps/sites don't support the .TIF format. So, this Go application does the following for me:

- Checks for new emails from the photo studio every 24 hours
- Visits the link in the email and downloads the .zip file
- Extracts the .TIF photos from the .zip file, and converts them into .PNGs
- Uploads the .TIF images to Google Drive for storage, and .PNG images to my AWS S3 bucket for sharing, such as on on my [personal website :)][portfolio-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Built With

- [![Go][Go]][Go-url]
- [![Gmail][Gmail]][Gcloud-url]
- [![Drive][Drive]][Gcloud-url]
- [![MongoDB][MongoDB]][MongoDB-url]
- [![AWSs3][AWSs3]][AWSs3-url]
- [![Fly][Fly]][Fly-url]

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->

## Getting Started

### Prerequisites

- [Go][Go-url] 1.23 or higher
- Database uri from [MongoDB][MongoDB-url]
- Client credentials from [Google Cloud][Gcloud-url]
- Bot token from [Discord][Discord-url]
- Access keys from [AWS S3][AWSs3-url]

### Installation

1. Clone the repo
   ```sh
   git clone https://github.com/ftrbnd/film-sync.git
   ```
2. Install Go modules
   ```sh
   go mod download
   ```
3. Start the local server
   ```sh
   make run
   ```

### Configuration

Create a `.env` file at the root and fill out the values:

```env
  FROM_EMAIL="noreply@wetransfer.com"
  REPLY_TO_EMAIL="giosalas25@gmail.com"

  MONGODB_URI=""

  CLIENT_ID=<some-value>.apps.googleusercontent.com"
  PROJECT_ID="film-sync"
  AUTH_URI="https://accounts.google.com/o/oauth2/auth"
  TOKEN_URI="https://oauth2.googleapis.com/token"
  AUTH_PROVIDER_X509_CERT_URL="https://www.googleapis.com/oauth2/v1/certs"
  CLIENT_SECRET=""
  REDIRECT_URI="" # ex: http://localhost:3001/auth or https://deployed-url/auth

  DISCORD_USER_ID="" # User ID to send messages to
  DISCORD_TOKEN=""

  AWS_ACCESS_KEY_ID=""
  AWS_SECRET_ACCESS_KEY=""
  AWS_BUCKET_NAME="" # ex: my-film-photos
  AWS_REGION="" # ex: us-west-1

  PORT=3001
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- USAGE EXAMPLES -->

## Usage

This is currently meant to server only one user, so those interested should clone the repo and deploy the app themselves

### Authentication

When the app first starts, a Discord message will be sent to the user asking them to sign in with their Google account. Following the sent link and signing in is all that's need to set up credentials

[![Auth Screenshot][auth-screenshot]](https://i.imgur.com/D845zqx.png)

### Upload Complete

Once all photos from a zip file have been successfully uploaded, the Discord bot will sent a confirmation message:
[![Success Screenshot][success-screenshot]](https://i.imgur.com/xYKpKAX.png)

<!-- CONTRIBUTING -->

## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTACT -->

## Contact

Giovanni Salas - [@finalcalI](https://twitter.com/finalcali) - giosalas25@gmail.com

Project Link: [https://github.com/ftrbnd/film-sync](https://github.com/ftrbnd/film-sync)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->

[Go]: https://img.shields.io/badge/go-00ADD8?style=for-the-badge&logo=go&logoColor=white
[Go-url]: https://go.dev
[Gmail]: https://img.shields.io/badge/gmail-EA4335?style=for-the-badge&logo=gmail&logoColor=white
[Drive]: https://img.shields.io/badge/google%20drive-4285F4?style=for-the-badge&logo=googledrive&logoColor=white
[Gcloud-url]: https://cloud.google.com
[MongoDB]: https://img.shields.io/badge/mongodb-47A248?style=for-the-badge&logo=mongodb&logoColor=white
[MongoDB-url]: https://mongodb.com
[Discord-url]: https://discord.com/developers/applications
[AWSs3]: https://img.shields.io/badge/amazon%20s3-569A31?style=for-the-badge&logo=amazons3&logoColor=white
[AWSs3-url]: https://aws.amazon.com/s3/
[Fly]: https://img.shields.io/badge/fly.io-24175B?style=for-the-badge&logo=flydotio&logoColor=white
[Fly-url]: https://fly.io
[portfolio-url]: https://giosalad.dev
[auth-screenshot]: https://i.imgur.com/D845zqx.png
[success-screenshot]: https://i.imgur.com/xYKpKAX.png