# 🛰️ Babel Agent

(This project is under Babel Foundation initiative. You can read the manifest [here](https://github.com/margostino/babel-foundation))

A daemon process that detects changes in local Babel data, pushes changes to the remote repository, and updates metadata and indexing.

<p align="center">
  <img src="https://github.com/margostino/babel-foundation/blob/master/assets/babel-architecture.png?raw=true" alt="Babel Foundation Architecture"/>
</p>

## Features

- **Change Detection**: Monitors local Babel data for changes and pushes updates to the remote repository.
- **Indexing and Metadata Updates**: Regularly updates indexing and metadata.
- **Customizable Interval**: Default interval of 30 seconds, configurable as needed.

### Requirements

```bash
ssh-add ~/.ssh/id_rsa
```
