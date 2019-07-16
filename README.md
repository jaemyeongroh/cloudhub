# SCMP (Snet Cloud Monitoring Portal)

## Intro.
**SCMP** is **another branch** of an open-source web application derived from Influxdata's **_Chronograf_** written in Go and React.js that provides the tools to visualize your monitoring data and easily create alerting and automation rules.

Therefore, SCMP will be enhanced by adding our direction, such as automation of configuration management and monitoring over Clouds.

SCMP has been started with _Chronograf_ version **1.7.11**.

### Using the basic common features
For more information of the basic common features between _Chronograf_ 1.7.11 and SCMP, dependencies and using guides like the way of TICK Script or Flux queries, refer to the following link.<br>
[Github for _chronograf_](https://github.com/influxdata/chronograf/blob/master/README.md)<br>
[Documents for TICK Stack](https://docs.influxdata.com/)

### Test Environment
For running this project, maybe, you shoud get the environment for test data and composition like telegraf, kapacity and influxdb.
This **_Sandbox_** provided by _Influxdata_ will help to do.<br>
[Download Sandbox](https://github.com/influxdata/sandbox)

## Key Differences against _Chronograf_ at this point.
* Compose of directories.
  * Divide as backend and frontend.
* More easy debugging environment support without proxy server by node.js.
* For more Dev. Env., provide **Visual Studio Code** Env. including the setting.json and launch.json.
* Window build & run Env. support.

## Setting in VSCode
* Add the followings into **_setting.json_** to **User Setting** namespace.
```
{
  "terminal.integrated.shell.windows": "C:\\Program Files\\Git\\bin\\bash.exe",
  "terminal.integrated.rightClickBehavior": "default",
  "terminal.explorerKind": "external",
  "terminal.integrated.copyOnSelection": true,
  "terminal.integrated.scrollback": 10000,
  "breadcrumbs.enabled": true,
  "editor.renderControlCharacters": true,
  "editor.largeFileOptimizations": false,
  "editor.formatOnSave": true,
  "editor.renderWhitespace": "none",
  "workbench.startupEditor": "newUntitledFile",
  "explorer.confirmDelete": false,
  "explorer.confirmDragAndDrop": false,
  "files.eol": "\n",
  "go.formatTool": "goimports",
  "go.lintOnSave": "package",
  "prettier.singleQuote": true,
  "prettier.bracketSpacing": false,
  "prettier.semi": false,
  "prettier.trailingComma": "es5",
  "eslint.alwaysShowStatus": true,
  "tslint.jsEnable": true,
  "[jsonc]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "[javascript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "debug.showInStatusBar": "always",
  "[typescriptreact]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "[json]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  },
  "debug.toolBarLocation": "docked",
  "[typescript]": {
    "editor.defaultFormatter": "esbenp.prettier-vscode"
  }
}
```

## How to build
### Getting the source code from github.
[If you're on Windows, run "Git Bash" and] type the followings.
```
# If you're on Windows, run "Git Bash" and type the followings.

$ go get github.com/snetsystems/cmp
$ cd $GOPATH/src/github.com/snetsystems/cmp
$ make
```

If well done, you can see the binary.
```
$ cd backend/cmd/cmp
$ ls -l
total 28072
...
-rwxr-xr-x 1 Snetsystems 197121 28610048 Jul 15 09:09 scmp
```
Once run scmp, 8888 port will be listened.
```
$ ./scmp
```
You can see the SCMP UI via browser: http://localhost:8888 

## How to debug via VSCode for Development.
For your convenience, make "_.code-workspace_" for VSCode in the your snetsystems folder.
```
$ cd $GOPATH/src/github.com/snetsystems/
$ cat snet.code-workspace
{
  "folders": [
    {
      "path": "cmp"
    }
  ],
  "settings": {
    "files.exclude": {}
  }
}
```
Run VSCode as above workspace.
```
$ code snet.code-workspace
```
Simply, select **"Debug SCMP"** and then run debug.<br>
Also, for UI debugging, select **"Launch Chrome"** and then run debug.

> We already prepared **"_.vscode/launch.json_"** and **"_.vscode/settings.json_"**
>> * Using **GO111MODULE**.
>>   * Not need a vendor directory anymore.
>> * Snetsystems Github login setting as a default.
>>   * You need to change to the Github's keys of your organization.
>>   * If you don't need to login, get rid of the login information.
>>     ```
>>      ...
>>      "args": [
>>        "-l=debug",
>>        "-d"
>>        // "--auth-duration=0",
>>        // "-t=74c1e9e2450886060b5bf736b935cd0bf960837f",
>>        // "--github-client-id=c170bbdba5cb2ea8c3e6",        
>>        // "--github-client-secret=55c35715b0e4eebab7edbdeef3081bf890e79d22"
>>      ],
>>      ...
>>     ```
