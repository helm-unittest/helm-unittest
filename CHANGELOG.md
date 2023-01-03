0.2.11 / 2023-01-03
===================
- Added lenghtEqual assertion to validate array counts (credits to: @lokkersp, resolves #190)
- Correct empty rendered manifest (credits to: @zifter)
- Correct subchart testing link (credits to: ludovicalarcon, resolves #185)
- Update documentation based on user questions
- Update packages to latest version

0.2.10 / 2022-11-07
===================
- Helm Unittest plugin is not available for non-root user (resolves #179)
- isSubset assertion to handle multiple keys (credits to: @iben12, resolves #162)
- Out Of Bounds array will result in a null value (resolves #167, #174)
- Additional debug option to validate failed tests with same expected and actual results (resolves #109, #180)
- Update documentation based on user questions
- Update packages to latest version

0.2.9 / 2022-09-24
==================
- Add JSON Schema for validating testsuite files (credits to: @armingerten, resolves #161)
- Support failedTemplate assert schema for valdiation errors (credits to: @rquino)
- Switch shell instead of bash to support other (credits to: @tewfik-ghariani)
- Correct loading appVersion (resolves #172)
- Update plugin to go 1.18
- Update documentation based on user questions
- Update packages to latest version

0.2.8 / 2021-11-01
==================
- Add support for Macosx arm64 (credits to: @svobol13)
- Add support for new assertion, containsDocument (credits to: @cyrus-mc)
- Add releasename validation (credits to: @jainbhavya65)
- Fixed reloading the project to prevent unwanted side-effects  (credits to: @armingerten, resolves #111)
- Fixed pre-processing rendered data before comparison (credits to: @wenhulove333)
- Update packages to latest version

0.2.7 / 2021-07-26
==================
- Added samples for contains mapping (resolves #107)
- Improved errorhandling, show complete error on failure, when failed_template validator is not used (resolves #109)
- Fixed import-values (credits to: @rquinio1A, resolves #115)
- Added samples for subsubcharts and global values (resolves #114)
- Fix small documentation improvements and corrections (credits to: @jglick, @krichter722, @craig-mcmahon)
- Update packages to lates version

0.2.6 / 2021-03-31
==================
- Add support for list of templates on tests (credits to: @stevelipinski)
- Add support for failfast (resolves #84)
- Add support for values files in testsuite (resolves #91)
- Add support for values files in path (resolves #92)
- Add support for strict validation of testsuites (resolves #80, #94)
- Fix contains assert validation, when count is used (resolves #98)
- Fix small documentation corrections (credits to: @mik-laj, @Michael03, @SaffatHasan)
- Update packages to lates version
- Added Frequently Asked Questions

0.2.5 / 2020-11-17 
==================
- Restructure solution to align more on go structure. (resolves #65)
- Fix improved validation for matchRegEx assertions (resolves #66)
- Feature add chart version override (resolves #67)
- Fix render only templates, selected in the test (resolves #68)
- Fix improved error filter when failed or required function is used in helm3 partial template (resolves #70)
- Fix wget installation (resolves #77)
- Fix sha256sum on CentOs (resolves #78)
- Update packages to lates version

0.2.4 / 2020-09-11
==================
- Fix resetting chart dependencies with conditions when running multiple tests in a testsuite (resolves #60 )
- Fix automatic publishing docker image after distribution (resolves #33 )

0.2.3 / 2020-08-31
==================
- Auto upload latest plugin version with a combination of helm clients (resolves #33 )
- Add support for setting Release on test suite level
- Add support for setting Capabilities, also on test suite level (resolves #36 ).
- Fix missing file assertion (resolves #39 , resolves #53 ).

0.2.2 / 2020-08-21
==================
- Add overriding capabilities within the testsuite (resolves #36).

0.2.1 / 2020-07-27
==================
- Add sha256sum validation when shasum is not available (resolves #35).
- Ignore validation when both sha tools are not available.
- Update dependencies to latest compatible versions
- Update installation script with sha256sum (resolves #38).
- Update 0.2.1 release (resolves #40)

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
