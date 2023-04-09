# CLI module

## Usage

`./example-module <action>` + protobuf data sent file at $CG_MODULE_ACTION_DATA_FILE.

## Responsibility

Everything that is language specific: library management, files (except `LICENSE`, `README`, `.codegame.json`, `.git` (but appending to `.gitignore`)),
type definitions (using `cge-parser`), running, building, …

## Actions

### status (required)

Returns supported actions, versions, etc.

```jsonc
{
  "actions": ["status", "create", "update", "run", "build"],
  "library_versions": ["0.1", "0.2"],
  "application_types": ["client", "server"]
}
```

### create

Creates a new project of the specified type.

#### client

Creates a new game client. This includes integration with the client library for the language,
a functioning template with a main file and language specific metadata files like package.json
or similar and wrappers around the library to make its usage easier including setting the game URL with
the CG_GAME_URL environment variable.

#### server

Creates a new game server. This includes integration with the server library of the language and a functioning template,
which implements common logic like starting the server and providing a game class.

### update

Updates the library to the specified version and all other dependencies used by the project to
their newest compatible version. Additionally all wrappers and type definitions are regenerated.

### run

Runs the project with the specified command line arguments and the CG_GAME_URL environment variable set
to the URL specified in the `.codegame.json` file.

### build

Builds the projects and injects the URL specified in the `.codegame.json` file,
which makes the CG_GAME_URL environment variable optional.
