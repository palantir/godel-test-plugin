godel-test-plugin
=================
godel-test-plugin is a gödel plugin that runs Go tests and provides the ability to group Go tests. It also provides the 
ability to write the output of Go tests as JUnit XML.

Plugin Tasks
------------
godel-test-plugin provides the following tasks:

* `test`: runs the tests for a project as defined by the configuration.

Tags
----
The configuration allows "tags" to be specified for tests. A tag is a name for a set of packages. The tests for a tag
are specified as matchers that match packages based on their path or name. By default, the `test` task runs on all
packages (except for those excluded by `exclude`). If tags are specified, then only the tests for the packages matching 
the tags are run. The `all` tag matches all packages that are part of any tag (that is, it is the union of all defined
tags). The `none` tag matches all packages that are not part of any defined tag. Any packages that are specified as
excluded are always excluded (regardless of the tag parameter).
