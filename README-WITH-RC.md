# Changelog

### [0.12.0-rc.1](https://github.com/ESG-USA/Auklet-Client-C/tree/0.12.0-rc.1)

**Implemented enhancements:**

- APM-1564 Transition from checksum to release [#75](https://github.com/ESG-USA/Auklet-Client-C/pull/75) ([kdsch](https://github.com/kdsch))
- APM-1596 Customer Defined Version in C/C++ Client/Agent (and a bugfix) [#74](https://github.com/ESG-USA/Auklet-Client-C/pull/74) ([kdsch](https://github.com/kdsch))
- APM-1415  C/C++ Client ASIL B Compliance [#65](https://github.com/ESG-USA/Auklet-Client-C/pull/65) ([kdsch](https://github.com/kdsch))

**Fixed bugs:**

- Empty String Rather than Null For No User Defined Version [#77](https://github.com/ESG-USA/Auklet-Client-C/pull/77) ([kdsch](https://github.com/kdsch))

**DevOps changes:**

- Add gofmt hook [#71](https://github.com/ESG-USA/Auklet-Client-C/pull/71) ([rjenkinsjr](https://github.com/rjenkinsjr))

## [0.11.0](https://github.com/ESG-USA/Auklet-Client-C/tree/0.11.0)

### [0.11.0-rc.2](https://github.com/ESG-USA/Auklet-Client-C/tree/0.11.0-rc.2)

**Fixed bugs:**

- cmd/client: let API determine emission period [#69](https://github.com/ESG-USA/Auklet-Client-C/pull/69) ([kdsch](https://github.com/kdsch))
- cmd/client: change emission period to 10 seconds [#68](https://github.com/ESG-USA/Auklet-Client-C/pull/68) ([kdsch](https://github.com/kdsch))

### [0.11.0-rc.1](https://github.com/ESG-USA/Auklet-Client-C/tree/0.11.0-rc.1)

**Implemented enhancements:**

- APM-1508  Transition MQTT Client ID From Device ID to Client ID [#66](https://github.com/ESG-USA/Auklet-Client-C/pull/66) ([kdsch](https://github.com/kdsch))
- APM-1483  C/C++ Client Device Registration [#64](https://github.com/ESG-USA/Auklet-Client-C/pull/64) ([kdsch](https://github.com/kdsch))
- APM-1415  C/C++ Client ASIL B Compliance [#62](https://github.com/ESG-USA/Auklet-Client-C/pull/62) ([kdsch](https://github.com/kdsch))
- APM-1432 C/C++ Package Version Information [#59](https://github.com/ESG-USA/Auklet-Client-C/pull/59) ([kdsch](https://github.com/kdsch))
- Use MessagePack Encoding for Transport [#58](https://github.com/ESG-USA/Auklet-Client-C/pull/58) ([kdsch](https://github.com/kdsch))
- APM-1415  C/C++ Client ASIL B Compliance [#57](https://github.com/ESG-USA/Auklet-Client-C/pull/57) ([kdsch](https://github.com/kdsch))
- APM-1359 Data Transmission Optimizations in C/C++ Client [#51](https://github.com/ESG-USA/Auklet-Client-C/pull/51) ([kdsch](https://github.com/kdsch))

**Fixed bugs:**

- .devops: fix incorrect package path [#63](https://github.com/ESG-USA/Auklet-Client-C/pull/63) ([kdsch](https://github.com/kdsch))
- schema: convert agent logs to raw messages [#56](https://github.com/ESG-USA/Auklet-Client-C/pull/56) ([kdsch](https://github.com/kdsch))

**DevOps changes:**

- Generalize gathering of core Golang licenses [#61](https://github.com/ESG-USA/Auklet-Client-C/pull/61) ([rjenkinsjr](https://github.com/rjenkinsjr))
- APM-1415  C/C++ Client ASIL B Compliance (Use Code Climate reporter) [#60](https://github.com/ESG-USA/Auklet-Client-C/pull/60) ([kdsch](https://github.com/kdsch))
- Fix some missing license texts [#55](https://github.com/ESG-USA/Auklet-Client-C/pull/55) ([rjenkinsjr](https://github.com/rjenkinsjr))

## [0.10.0](https://github.com/ESG-USA/Auklet-Client-C/tree/0.10.0)

### [0.10.0-rc.1](https://github.com/ESG-USA/Auklet-Client-C/tree/0.10.0-rc.1)

**Implemented enhancements:**

- APM-1353 Use generic terms instead of names of specific technologies [#48](https://github.com/ESG-USA/Auklet-Client-C/pull/48) ([kdsch](https://github.com/kdsch))
- APM-1333 C/C++ Agent Logging [#43](https://github.com/ESG-USA/Auklet-Client-C/pull/43) ([kdsch](https://github.com/kdsch))

**Fixed bugs:**

- message: don't doubly close queue output channel [#47](https://github.com/ESG-USA/Auklet-Client-C/pull/47) ([kdsch](https://github.com/kdsch))

**DevOps changes:**

- Fix CircleCI Docker image [#50](https://github.com/ESG-USA/Auklet-Client-C/pull/50) ([rjenkinsjr](https://github.com/rjenkinsjr))

## [0.9.0](https://github.com/ESG-USA/Auklet-Client-C/tree/0.9.0)

### [0.9.0-rc.1](https://github.com/ESG-USA/Auklet-Client-C/tree/0.9.0-rc.1)

**Implemented enhancements:**

- License under Apache 2.0 / harvest dependency licenses [#41](https://github.com/ESG-USA/Auklet-Client-C/pull/41) ([rjenkinsjr](https://github.com/rjenkinsjr))
- APM-1320 Allow Console Logs in Production Releases of Client [#32](https://github.com/ESG-USA/Auklet-Client-C/pull/32) ([kdsch](https://github.com/kdsch))

**Fixed bugs:**

- APM-1335 No C data on Staging [#42](https://github.com/ESG-USA/Auklet-Client-C/pull/42) ([kdsch](https://github.com/kdsch))

**DevOps changes:**

- Push prod branch to aukletio [#44](https://github.com/ESG-USA/Auklet-Client-C/pull/44) ([rjenkinsjr](https://github.com/rjenkinsjr))
- Improve WhiteSource integration [#40](https://github.com/ESG-USA/Auklet-Client-C/pull/40) ([rjenkinsjr](https://github.com/rjenkinsjr))
- Add WhiteSource integration [#38](https://github.com/ESG-USA/Auklet-Client-C/pull/38) ([rjenkinsjr](https://github.com/rjenkinsjr))
- Fix prod PR update script [#37](https://github.com/ESG-USA/Auklet-Client-C/pull/37) ([rjenkinsjr](https://github.com/rjenkinsjr))
- Fix changelog generation syntax [#36](https://github.com/ESG-USA/Auklet-Client-C/pull/36) ([rjenkinsjr](https://github.com/rjenkinsjr))
- TS-419: Stop using GitHub API for gathering commit lists [#35](https://github.com/ESG-USA/Auklet-Client-C/pull/35) ([rjenkinsjr](https://github.com/rjenkinsjr))
- TS-417: update prod release PR after QA release finishes [#34](https://github.com/ESG-USA/Auklet-Client-C/pull/34) ([rjenkinsjr](https://github.com/rjenkinsjr))
- APM-1329: Fix GitHub API abuse rate limits [#33](https://github.com/ESG-USA/Auklet-Client-C/pull/33) ([rjenkinsjr](https://github.com/rjenkinsjr))

## [0.8.0](https://github.com/ESG-USA/Auklet-Client-C/tree/0.8.0)

### [0.8.0-rc.1](https://github.com/ESG-USA/Auklet-Client-C/tree/0.8.0-rc.1)

**Implemented enhancements:**

- APM-1235 Local Data Storage, APM-1234 Data Upload Limit [#22](https://github.com/ESG-USA/Auklet-Client-C/pull/22) ([kdsch](https://github.com/kdsch))

**Fixed bugs:**

- api: do not use DisallowUnknownFields [#30](https://github.com/ESG-USA/Auklet-Client-C/pull/30) ([kdsch](https://github.com/kdsch))
- api: return if GET request returns an error [#28](https://github.com/ESG-USA/Auklet-Client-C/pull/28) ([kdsch](https://github.com/kdsch))
- Troubleshooting bugs in Docker for VDAS [#27](https://github.com/ESG-USA/Auklet-Client-C/pull/27) ([kdsch](https://github.com/kdsch))

## [0.7.0](https://github.com/ESG-USA/Auklet-Client-C/tree/0.7.0)

### [0.7.0-rc.1](https://github.com/ESG-USA/Auklet-Client-C/tree/0.7.0-rc.1)

**Implemented enhancements:**

- Accept one CA cert file [#24](https://github.com/ESG-USA/Auklet-Client-C/pull/24) ([kdsch](https://github.com/kdsch))
- APM-1276 Bidirectional Communication for Agent and Client [#23](https://github.com/ESG-USA/Auklet-Client-C/pull/23) ([kdsch](https://github.com/kdsch))

## [0.6.0](https://github.com/ESG-USA/Auklet-Client-C/tree/0.6.0)

### [0.6.0-rc.1](https://github.com/ESG-USA/Auklet-Client-C/tree/0.6.0-rc.1)

**Implemented enhancements:**

- device: change device metrics JSON fields [#17](https://github.com/ESG-USA/Auklet-Client-C/pull/17) ([kdsch](https://github.com/kdsch))
- APM-1215 Change C event's JSON [#16](https://github.com/ESG-USA/Auklet-Client-C/pull/16) ([kdsch](https://github.com/kdsch))
- APM-1215 Change C event's JSON [#15](https://github.com/ESG-USA/Auklet-Client-C/pull/15) ([kdsch](https://github.com/kdsch))
- APM-1172 APM-1215 Change Tree and Event Fields [#11](https://github.com/ESG-USA/Auklet-Client-C/pull/11) ([kdsch](https://github.com/kdsch))

## [0.5.0](https://github.com/ESG-USA/Auklet-Client-C/tree/0.5.0)

### [0.5.0-rc.1](https://github.com/ESG-USA/Auklet-Client-C/tree/0.5.0-rc.1)

**Implemented enhancements:**

- APM-1172 Change Tree Fields [#9](https://github.com/ESG-USA/Auklet-Client-C/pull/9) ([kdsch](https://github.com/kdsch))

**Fixed bugs:**

- Revert "APM-1172 Change Tree Fields" [#10](https://github.com/ESG-USA/Auklet-Client-C/pull/10) ([MZein1292](https://github.com/MZein1292))

**DevOps changes:**

- APM-1177: fix changelog generation [#12](https://github.com/ESG-USA/Auklet-Client-C/pull/12) ([rjenkinsjr](https://github.com/rjenkinsjr))

## [0.4.0](https://github.com/ESG-USA/Auklet-Client-C/tree/0.4.0)

### [0.4.0-rc.3](https://github.com/ESG-USA/Auklet-Client-C/tree/0.4.0-rc.3)

**Implemented enhancements:**

- APM-1125 Reorganize/distribute "docs" in Auklet-Agent-C repo [#8](https://github.com/ESG-USA/Auklet-Client-C/pull/8) ([kdsch](https://github.com/kdsch))
- APM-1134: Hardcode BASE_URL when not built locally [#7](https://github.com/ESG-USA/Auklet-Client-C/pull/7) ([rjenkinsjr](https://github.com/rjenkinsjr))
- APM-1091 Fetch Kafka parameters from API [#6](https://github.com/ESG-USA/Auklet-Client-C/pull/6) ([kdsch](https://github.com/kdsch))
- APM-1065 Refactoring Client Codebase [#3](https://github.com/ESG-USA/Auklet-Client-C/pull/3) ([kdsch](https://github.com/kdsch))
- APM-1090: Separate Auklet-Profiler-C into separate repos [#2](https://github.com/ESG-USA/Auklet-Client-C/pull/2) ([rjenkinsjr](https://github.com/rjenkinsjr))

**DevOps changes:**

- TS-409: Do not consider PRs not merged to HEAD [#4](https://github.com/ESG-USA/Auklet-Client-C/pull/4) ([rjenkinsjr](https://github.com/rjenkinsjr))
