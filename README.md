# gctl

CLI for working with Google Products Like gmail and drive

# Getting Started

Create an OAuth Client ID using the Google Developer's Console.
Download the JSON file and save it. Then configure the CLI to use it

```
gctl config set oAuthClientFile=/PATH/TO/YOUR/CLIENT.SECRET.json
```

Enable the Gmail API in the developers console for the project in which you created the OAuth Client ID.