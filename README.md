# msgraphbeat

Welcome to {Beat}.

Ensure that this folder is at the following location:
`${GOPATH}/src/github.com/marian-craciunescu/msgraphbeat`

## Getting Started with {Beat}

### Requirements

* [Golang](https://golang.org/dl/) 1.13

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

### Build

To build the binary for {Beat} run the command below. This will generate a binary
in the same directory with the name msgraphbeat.

```
make
```


### Prerquisite
In addition, you need a Microsoft application registered to access the Microsoft Graph, as
described in Authorization and the Microsoft Graph Security API. The following steps are a
summary of the procedures from this article. Note that any updates to the article may supersede
the steps presented here.
1. Create the application:

    a: Sign in to the Application Active Directory portal ( https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/Overview ) using your Microsoft account.
   
    b. Under *App Registration* choose  *New Registrations*. 

    c. Enter an application name, and choose Create.
    
    d. In the registration page for your app, copy and save the Application ID field. You need
it later to complete the configuration process.

    e. Under *Certificates and Secrets *, choose Generate New Client Secret. A new password is
displayed in the New password generated dialog.

    f. IMPORTANT: Copy the password. You need it later to complete the configuration
process and you will not be able to see the secret again.
    
    g. Under Api Permission, choose Add a Permission > Microsoft.Graph.
h. Under Application Permissions, add the permissions *SecurityEvents.Read.All*,
and *SecurityEvents.ReadWrite.All*. 
See the Microsoft Graph permissions reference for
more information about Graph's permission model. https://docs.microsoft.com/en-us/graph/permissions-reference

    i. Enter http://localhost as the Redirect URL, and then choose Save.
2. Give Administrator consent to view Security data:

    a. Provide to your Microsoft Administrator account your Application ID and the Redirect
URI that you used in the previous steps. The organization’s Administrator (or other user
authorized to grant consent for organizational security resources) is required to grant
consent to the application.
    
    b. As the tenant Admin with Security Administrator privileges for your organization, open a
browser window and craft the following URL in the address bar:
https://login.microsoftonline.com/common/adminconsent?client_id=APPLICAT
ION_ID&state=12345
Where APPLICATION_ID is the application ID value from the App V2 registration portal,
which you can view after clicking on your application to view its properties.

    c.
After logging in, the tenant Admin is presented with a dialog similar to the following:
    d. When the tenant Admin agrees to this dialog, the administrator is granting consent for all
users of their organization to use this application.
F
or more details about the authorization flow, read the Authorization and the Microsoft Graph
Security API.


### Run

To run {Beat} with debugging output enabled, run:

```
./msgraphbeat -c msgraphbeat.yml -e -d "*"
```


### Test

To test {Beat}, run the following command:

```
make testsuite
```

alternatively:
```
make unit-tests
make system-tests
make integration-tests
make coverage-report
```

The test coverage is reported in the folder `./build/coverage/`

### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `fields.yml` by running the following command.

```
make update
```

## Packaging

The beat frameworks provides tools to crosscompile and package your beat for different platforms. This requires [docker](https://www.docker.com/) and vendoring as described above. To build packages of your beat, run the following command:

```
make release
```

This will fetch and create all images required for the build process. The whole process to finish can take several minutes.
