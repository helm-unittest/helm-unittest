0.3.5 / 2023-09-01
==================
- Fix  (resolves #)
- Fix failed_template to also work when empty documents are found (resolves #191)
- Fix failed_template multi colon handling (resolves #200)
- Fix glob all valid filenames (resolves #201)
- Update packages to latest patch versions
- Update documenation

0.3.4 / 2023-08-01
===================
- Fix only output JUnit error when tests are failed (resolves #154)
- Fix/Refactor containsDocument validation, handles strict validation when multiple documents are found (resolves #167, resolves #173)
- Fix schema definition types (resolves #174)
- Fix validation of required fields in suite (resolves #178)
- Remove GitHub API usage during instal (credits @raxod502-plaid, resolves #181)
- Enable suite-level set block (resolves #155)
- Update packages to latest patch versions
- Update documentation

0.3.3 / 2023-05-21
===================
- Fix template order which result in stable assertion validations (resolves #133)
- Fix negative containsDocument validation when an empty document is found (resolves #145)
- Fix template filter, to only load templates that are defined (resolves #153)
- Fix JUnit error output (resolves #154)
- Fix loading tpl files and yml files (resolves #158)
- Update examples to validate multiple templates (resolves #142, resolves #149)
- Update documentation to use values files in suite (resolves #155)
- Update packages to latest patch versions

0.3.2 / 2023-04-17
===================
- Fix tests not rendering when using $.Files.Get (resolves #135)
- Refactor IsNull to Exists and IsEmpty to IsNullOrEmpty (resolves #134)
- Update documentation based on user questions (resolves #129)

0.3.1 / 2023-04-10
===================
- Improvement JUnit export format (credits @steigr)
- Enable wildcard selection for test templates (resolves quintush/helm-unittest#173, quintush/helm-unittest#192)
- Update package name to align with github organisation (credits @mavimo, resolved #127)
- Fix set and values to be used simultaneously (resolves #124)
- Fix test suite code completion and make general available in https://www.schemastore.org (credits @armingerten, resolves quintush/helm-unittest#161)
- Stabelize template loading (resolves #123, quintush/helm-unittest#143, quintush/helm-unittest#205)
- Update documentation based on user questions
- Update packages to latest version
- Update docker containers to latest version

0.3.0 / 2023-01-30
===================
- Moved to origal repo (resolves quintush/helm-unittest#124)
- Add JsonPath (resolves quintush/helm-unittest#85, quintush/helm-unittest#158)
- Add Antonym to LengthEqual (resolves quintush/helm-unittest#197)
- Remove Helm 2 support
- Update documentation based on user questions
- Update packages to latest version

0.2.11 / 2023-01-03
===================
- Added lenghtEqual assertion to validate array counts (credits to: @lokkersp, resolves quintush/helm-unittest#190)
- Correct empty rendered manifest (credits to: @zifter)
- Correct subchart testing link (credits to: ludovicalarcon, resolves quintush/helm-unittest#185)
- Update documentation based on user questions
- Update packages to latest version

0.2.10 / 2022-11-07
===================
- Helm Unittest plugin is not available for non-root user (resolves quintush/helm-unittest#179)
- isSubset assertion to handle multiple keys (credits to: @iben12, resolves quintush/helm-unittest#162)
- Out Of Bounds array will result in a null value (resolves quintush/helm-unittest#167, quintush/helm-unittest#174)
- Additional debug option to validate failed tests with same expected and actual results (resolves quintush/helm-unittest#109, quintush/helm-unittest#180)
- Update documentation based on user questions
- Update packages to latest version

0.2.9 / 2022-09-24
==================
- Add JSON Schema for validating testsuite files (credits to: @armingerten, resolves quintush/helm-unittest#161)
- Support failedTemplate assert schema for valdiation errors (credits to: @rquino)
- Switch shell instead of bash to support other (credits to: @tewfik-ghariani)
- Correct loading appVersion (resolves quintush/helm-unittest#172)
- Update plugin to go 1.18
- Update documentation based on user questions
- Update packages to latest version

0.2.8 / 2021-11-01
==================
- Add support for Macosx arm64 (credits to: @svobol13)
- Add support for new assertion, containsDocument (credits to: @cyrus-mc)
- Add releasename validation (credits to: @jainbhavya65)
- Fixed reloading the project to prevent unwanted side-effects  (credits to: @armingerten, resolves quintush/helm-unittest#111)
- Fixed pre-processing rendered data before comparison (credits to: @wenhulove333)
- Update packages to latest version

0.2.7 / 2021-07-26
==================
- Added samples for contains mapping (resolves quintush/helm-unittest#107)
- Improved errorhandling, show complete error on failure, when failed_template validator is not used (resolves quintush/helm-unittest#109)
- Fixed import-values (credits to: @rquinio1A, resolves quintush/helm-unittest#115)
- Added samples for subsubcharts and global values (resolves quintush/helm-unittest#114)
- Fix small documentation improvements and corrections (credits to: @jglick, @krichter722, @craig-mcmahon)
- Update packages to lates version

0.2.6 / 2021-03-31
==================
- Add support for list of templates on tests (credits to: @stevelipinski)
- Add support for failfast (resolves quintush/helm-unittest#84)
- Add support for values files in testsuite (resolves quintush/helm-unittest#91)
- Add support for values files in path (resolves quintush/helm-unittest#92)
- Add support for strict validation of testsuites (resolves quintush/helm-unittest#80, quintush/helm-unittest#94)
- Fix contains assert validation, when count is used (resolves quintush/helm-unittest#98)
- Fix small documentation corrections (credits to: @mik-laj, @Michael03, @SaffatHasan)
- Update packages to lates version
- Added Frequently Asked Questions

0.2.5 / 2020-11-17 
==================
- Restructure solution to align more on go structure. (resolves quintush/helm-unittest#65)
- Fix improved validation for matchRegEx assertions (resolves quintush/helm-unittest#66)
- Feature add chart version override (resolves quintush/helm-unittest#67)
- Fix render only templates, selected in the test (resolves quintush/helm-unittest#68)
- Fix improved error filter when failed or required function is used in helm3 partial template (resolves quintush/helm-unittest#70)
- Fix wget installation (resolves quintush/helm-unittest#77)
- Fix sha256sum on CentOs (resolves quintush/helm-unittest#78)
- Update packages to lates version

0.2.4 / 2020-09-11
==================
- Fix resetting chart dependencies with conditions when running multiple tests in a testsuite (resolves quintush/helm-unittest#60)
- Fix automatic publishing docker image after distribution (resolves quintush/helm-unittest#33)

0.2.3 / 2020-08-31
==================
- Auto upload latest plugin version with a combination of helm clients (resolves quintush/helm-unittest#33)
- Add support for setting Release on test suite level
- Add support for setting Capabilities, also on test suite level (resolves quintush/helm-unittest#36).
- Fix missing file assertion (resolves quintush/helm-unittest#39 , resolves quintush/helm-unittest#53).

0.2.2 / 2020-08-21
==================
- Add overriding capabilities within the testsuite (resolves quintush/helm-unittest#36).

0.2.1 / 2020-07-27
==================
- Add sha256sum validation when shasum is not available (resolves quintush/helm-unittest#35).
- Ignore validation when both sha tools are not available.
- Update dependencies to latest compatible versions
- Update installation script with sha256sum (resolves #60).
- Update 0.2.1 release (resolves quintush/helm-unittest#40)

0.2.0 / 2020-04-20
==================
- having more assertions:
  - isSubset (resolve #6)
  - equalRaw (resolve #11)
  - matchRegexRaw (resolve #11)
  - matchSnapshotRaw (resolve #11)
  - contains, expanded with count value (resolve #52)
  - contains, expanded with any boolean (resolve #74)
  - failedTemplate (resolve #39, resolve #82)
- added support to validate multiple templates (resolve #38, resolve #54)
- added support to use checksum validation for release and install (resolve #60)
- added support to test dependent charts (resolve #65)
- fixed templates in subdirectories fail to load (fixed #44)
- bumped git modules (fixes #79, fixes #80)
- fixed support capabilities (refactered rendingen of the charts) (fixed #88)
- update to latest Helm2 library to support deepclone (fixes #96)
- improved download version, to download different arch and fully backwards compatible with older version (fixes #97)

0.1.8 / 2020-04-03
==================
- added jq syntax including test verifications (#95)

0.1.7 / 2020-04-02
==================
- added Helm V3 compatiblity (#87, #98)
- make install-binary.sh version aware (#97)

0.1.6 / 2019-10-14
==================
- added xml outputs JUnit, NUnit, XUnit and update project to use modules (#51, #78)

0.1.5 / 2019-04-09
==================
- update sprig (#72, #73)

0.1.4 / 2019-03-30
==================
- fix slash problem in windows (#70)
- add update plugin hook, enable `helm plugin update` (#69)

0.1.3 / 2019-03-29
==================
- use yaml.Decoder to parse multi doc manifest (#66)
- fix doc typo (#56, #63)
- upgrade sprig and helm (#49)
- fix static linking of building (#46)
- enhance install script (#43)
- standard dockerfile for running (#42)

0.1.2 / 2018-03-29
==================
- feature: recursively find test suite files along dependencies in `charts`
- fix: absolute value file path in TestJob.Values
- doc: fix `isAPIVersion` typo
- upgrade helm to v2.8.2
- more robust tests (of the plugin)
