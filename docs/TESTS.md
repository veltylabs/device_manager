# Pruebas

| Test | Regla |
|---|---|
| [TestMount_MatchesDefinition](../tests/conformance_test.go) | Mount — Matches Definition |
| [TestTenant_ReadOtherTenantIsNotFound](../tests/conformance_test.go) | Tenant — Read Other Tenant Is Not Found |
| [TestDevice_CreateRejectsClientTenant](../tests/conformance_test.go) | Device — Create Rejects Client Tenant |
| [TestDevice_DeleteRemovesItsInterfaces](../tests/conformance_test.go) | Device — Delete Removes Its Interfaces |
| [TestDemo_Valid](../tests/demo_test.go) | Demo — Valid |
| [TestNew_RequiresIDs](../tests/device_test.go) | New — Requires IDs |
| [TestNew_RequiresTenantID](../tests/device_test.go) | New — Requires Tenant ID |
| [TestCreateDevice_HappyPath](../tests/device_test.go) | Create Device — Happy Path |
| [TestCreateDevice_InvalidType](../tests/device_test.go) | Create Device — Invalid Type |
| [TestCreateDevice_DuplicateIPSameTenant](../tests/device_test.go) | Create Device — Duplicate IPSame Tenant |
| [TestCreateDevice_SameIPDifferentTenants_Allowed](../tests/device_test.go) | Create Device — Same IPDifferent Tenants — Allowed |
| [TestUpdateDevice_NotFound](../tests/device_test.go) | Update Device — Not Found |
| [TestDeactivateDevice](../tests/device_test.go) | Deactivate Device |
| [TestDeleteDevice](../tests/device_test.go) | Delete Device |
| [TestListDevices_FilterByTypeAndActive](../tests/device_test.go) | List Devices — Filter By Type And Active |
| [TestCreateDevice_PublishesEvent](../tests/device_test.go) | Create Device — Publishes Event |
| [TestSentinelErrorsStrings](../tests/errors_test.go) | Sentinel Errors Strings |
| [TestInterface_AutoAssignsLowestFreeIP](../tests/interface_test.go) | Interface — Auto Assigns Lowest Free IP |
| [TestInterface_ZoneFull](../tests/interface_test.go) | Interface — Zone Full |
| [TestInterface_NoZoneNeedsExplicitIP](../tests/interface_test.go) | Interface — No Zone Needs Explicit IP |
| [TestInterface_IPOutsideZone](../tests/interface_test.go) | Interface — IPOutside Zone |
| [TestInterface_MACRules](../tests/interface_test.go) | Interface — MACRules |
| [TestInterface_DeleteDeviceDeletesInterfaces](../tests/interface_test.go) | Interface — Delete Device Deletes Interfaces |
| [TestInterface_FindByIPThroughInterface](../tests/interface_test.go) | Interface — Find By IPThrough Interface |
| [TestItemsDeclaredOnFields](../tests/item_test.go) | Items Declared On Fields |
| [TestLegacyDeviceIP_OneGuardedStatement](../tests/legacy_ip_test.go) | El paso legado es una sola sentencia, protegida por la existencia de device.ip: |
| [TestIPLocator](../tests/locator_test.go) | IPLocator |
| [TestIPLocator_CanonicalSpelling](../tests/locator_test.go) | La IP con la que se guarda un equipo y la IP que auth observa al iniciar sesión |
| [TestMigrate_UpToDate_OnlyEnsuresTables](../tests/migrate_test.go) | Branch "listing works, nothing missing": only the idempotent CREATE TABLE per table, in order. |
| [TestMigrate_NewField_AddsOnlyThatColumn](../tests/migrate_test.go) | Branch "field missing from the table": exactly one ADD COLUMN, for that field of that table. |
| [TestMigrate_IntrospectionFails_ReturnsErrorAndStops](../tests/migrate_test.go) | Branch "listing the columns failed": the error comes back and no later table is touched. |
| [TestNetworkHosts_OnlyActiveDevicesInterfacesWithMAC](../tests/network_test.go) | Network Hosts — Only Active Devices Interfaces With MAC |
| [TestNetworkHosts_ImportHosts](../tests/network_test.go) | Network Hosts — Import Hosts |
| [TestFormWidgets](../tests/network_test.go) | Los formularios que construye crudview deben mostrar exactamente los campos editables. |
| [TestScreenSave](../tests/screen_save_test.go) | Screen Save |
| [TestScreen_TenantFromRequestIsRejected](../tests/screen_save_test.go) | Screen — Tenant From Request Is Rejected |
| [TestSeedLoad](../tests/seed_test.go) | Seed Load |
| [TestTenantIsolation](../tests/tenant_test.go) | Tenant Isolation |
| [TestUI_ScreensAreDescribedWithTheirOwnIcon](../tests/ui_described_test.go) | Las tres pantallas se explican solas, validan según view.Screen y cada una tiene su propio dibujo y permiso. |
| [TestZone_SaveListDelete](../tests/zone_test.go) | Zone — Save List Delete |
| [TestZone_InvalidRange](../tests/zone_test.go) | Zone — Invalid Range |
| [TestZone_Overlap](../tests/zone_test.go) | Zone — Overlap |
| [TestZone_DeleteInUse](../tests/zone_test.go) | Zone — Delete In Use |
| [TestZone_TenantIsolation](../tests/zone_test.go) | Zone — Tenant Isolation |
