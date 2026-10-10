# 系统设置模块接口级 E2E 测试
#
# 安全约定：本脚本只创建 / 删除带 "ZZE2E" 前缀的测试数据，
# 绝不按通用编码（如 E001、D001）清理，避免误删真实业务数据。
$ErrorActionPreference = 'Continue'
$base = 'http://127.0.0.1:88'
$script:pass = 0
$script:fail = 0
$script:removed = 0

function Check($name, $cond, $detail = '') {
  if ($cond) { $script:pass++; Write-Host ("PASS  " + $name) }
  else { $script:fail++; Write-Host ("FAIL  " + $name + "   => " + $detail) }
}

function Api {
  param([string]$Method, [string]$Path, $Json, [string]$Token)
  $headers = @{}
  if ($Token) { $headers['Authorization'] = "Bearer $Token" }
  $req = @{ Uri = ($base + $Path); Method = $Method; Headers = $headers; TimeoutSec = 60 }
  if ($null -ne $Json) {
    $req['Body'] = $Json
    $req['ContentType'] = 'application/json; charset=utf-8'
  }
  return Invoke-RestMethod @req
}

function Body($obj) { return ($obj | ConvertTo-Json -Depth 10 -Compress) }

function MenuIdByCode($tree, $code) {
  foreach ($n in $tree) {
    if ($n.menuCode -eq $code) { return $n.id }
    if ($null -ne $n.children) {
      $r = MenuIdByCode $n.children $code
      if ($null -ne $r) { return $r }
    }
  }
  return $null
}

function RouteNames($routes) {
  # 用逗号包一层，避免只有一个元素时 PowerShell 把数组拆成字符串
  $result = [string[]]@($routes | ForEach-Object { $_.name })
  return ,$result
}

$script:testDeptCodes = @('ZZE2E_D001', 'ZZE2E_D002', 'ZZE2E_D003')
$script:testEmpCodes = @('ZZE2E_E001', 'ZZE2E_E002', 'ZZE2E_E003')
$script:testUsers = @('zze2e_user')

function CleanDeptTree($nodes, $token) {
  foreach ($n in $nodes) {
    if ($null -ne $n.children -and @($n.children).Count -gt 0) { CleanDeptTree $n.children $token; continue }
    if ($script:testDeptCodes -contains $n.deptCode) {
      $null = Api DELETE ("/api/system/departments/" + $n.id) $null $token
      $script:removed++
      Write-Host ("  cleanup dept " + $n.deptCode)
    }
  }
}

Write-Host "`n=== 1. 登录与身份 ==="
$login = Api POST '/auth/login' (Body @{ userName = 'Soybean'; password = '123456' }) $null
Check 'admin 账号 Soybean/123456 登录成功' ($login.code -eq '0000') ($login | ConvertTo-Json -Compress)
$adminToken = $login.data.token
Check '登录返回 token' ([bool]$adminToken)

$info = Api GET '/auth/getUserInfo' $null $adminToken
Check 'getUserInfo 返回角色 R_SUPER' ($info.data.roles -contains 'R_SUPER') ($info | ConvertTo-Json -Compress)

$bad = Api POST '/auth/login' (Body @{ userName = 'Soybean'; password = 'wrong-pwd' }) $null
Check '错误密码被拒绝' ($bad.code -ne '0000')

$noauth = Api GET '/api/system/departments' $null $null
Check '未带令牌访问受保护接口被拒' ($noauth.code -eq '8888') ($noauth | ConvertTo-Json -Compress)

Write-Host "`n=== 1b. 清理上次中断留下的测试数据 ==="
$elx = Api GET '/api/system/employees?page=1&pageSize=200' $null $adminToken
foreach ($e in @($elx.data.list)) {
  if ($script:testEmpCodes -contains $e.empCode) {
    $null = Api DELETE ("/api/system/employees/" + $e.id) $null $adminToken
    Write-Host ("  cleanup employee " + $e.empCode)
  }
}
$ulx = Api GET '/api/system/users?page=1&pageSize=200' $null $adminToken
foreach ($u in @($ulx.data.list)) {
  if ($script:testUsers -contains $u.username) {
    $null = Api DELETE ("/api/system/users/" + $u.id) $null $adminToken
    Write-Host ("  cleanup user " + $u.username)
  }
}
for ($i = 0; $i -lt 6; $i++) {
  $dtx = Api GET '/api/system/departments' $null $adminToken
  $script:removed = 0
  CleanDeptTree $dtx.data.list $adminToken
  if ($script:removed -eq 0) { break }
}

Write-Host "`n=== 2. 菜单树接口 ==="
$mt = Api GET '/api/system/menuTree' $null $adminToken
Check 'menuTree 接口成功' ($mt.code -eq '0000') ($mt | ConvertTo-Json -Compress)
$menuRoots = $mt.data.list
Check 'menuTree 只有一个顶级目录 setting' (@($menuRoots).Count -eq 1 -and $menuRoots[0].menuCode -eq 'setting') ($menuRoots | ConvertTo-Json -Compress)
Check 'menuTree setting 下有 3 个子菜单' (@($menuRoots[0].children).Count -eq 3) (@($menuRoots[0].children) | ConvertTo-Json -Compress)

$idSetting = MenuIdByCode $menuRoots 'setting'
$idDept = MenuIdByCode $menuRoots 'setting_department'
$idEmp = MenuIdByCode $menuRoots 'setting_employee'
$idUser = MenuIdByCode $menuRoots 'setting_user'
Check '四个菜单ID均可解析' ([bool]$idSetting -and [bool]$idDept -and [bool]$idEmp -and [bool]$idUser) "$idSetting/$idDept/$idEmp/$idUser"

Write-Host "`n=== 3. 当前用户动态路由（admin 全量） ==="
$ur = Api GET '/route/getUserRoutes' $null $adminToken
Check 'getUserRoutes 成功' ($ur.code -eq '0000') ($ur | ConvertTo-Json -Compress)
$names = RouteNames $ur.data.routes
Check 'admin 路由含 home/init/setting' (($names -contains 'home') -and ($names -contains 'init') -and ($names -contains 'setting')) ($names -join ',')
Check 'home 为默认落地路由' ($ur.data.home -eq 'home') $ur.data.home
$settingNode = $ur.data.routes | Where-Object { $_.name -eq 'setting' }
Check 'setting 节点使用 layout.base' ($settingNode.component -eq 'layout.base') $settingNode.component
$childNames = RouteNames $settingNode.children
Check 'setting 子路由为 3 个设置页' (@($childNames).Count -eq 3) ($childNames -join ',')
Check '子路由 component 为 view.setting_*' (@($settingNode.children | Where-Object { $_.component -notlike 'view.setting_*' }).Count -eq 0) (@($settingNode.children | ForEach-Object { $_.component }) -join ',')
$homeNode = $ur.data.routes | Where-Object { $_.name -eq 'home' }
Check '单级路由 home 使用 layout.base$view.home' ($homeNode.component -eq 'layout.base$view.home') $homeNode.component
$initNode = $ur.data.routes | Where-Object { $_.name -eq 'init' }
Check 'init 使用空白布局且隐藏于菜单' ($initNode.component -eq 'layout.blank$view.init' -and $initNode.meta.hideInMenu) ($initNode | ConvertTo-Json -Compress)
Check 'setting 节点带 redirect 指向首个页面' ($settingNode.redirect -eq '/setting/department') $settingNode.redirect

$ie1 = Api GET '/route/isRouteExist?routeName=setting_user' $null $adminToken
Check 'isRouteExist(setting_user)=true' ($ie1.data -eq $true) ($ie1 | ConvertTo-Json -Compress)
$ie2 = Api GET '/route/isRouteExist?routeName=not_exist_route' $null $adminToken
Check 'isRouteExist(not_exist_route)=false' ($ie2.data -eq $false) ($ie2 | ConvertTo-Json -Compress)

Write-Host "`n=== 4. 部门 CRUD ==="
$d1 = Api POST '/api/system/departments' (Body @{ id = 0; deptCode = 'ZZE2E_D001'; deptName = '财务部'; parentId = 0; leaderId = 0; sortOrder = 1; isActive = 1 }) $adminToken
Check '新增顶级部门成功' ($d1.code -eq '0000') ($d1 | ConvertTo-Json -Compress)
$deptId1 = $d1.data.id

$d2 = Api POST '/api/system/departments' (Body @{ id = 0; deptCode = 'ZZE2E_D002'; deptName = '会计组'; parentId = $deptId1; leaderId = 0; sortOrder = 1; isActive = 1 }) $adminToken
Check '新增子部门成功' ($d2.code -eq '0000') ($d2 | ConvertTo-Json -Compress)
$deptId2 = $d2.data.id

$dup = Api POST '/api/system/departments' (Body @{ id = 0; deptCode = 'ZZE2E_D001'; deptName = '重复编码'; parentId = 0; leaderId = 0; sortOrder = 9; isActive = 1 }) $adminToken
Check '重复部门编码被拒绝' ($dup.code -ne '0000') ($dup | ConvertTo-Json -Compress)

$dt = Api GET '/api/system/departments' $null $adminToken
$root = $dt.data.list | Where-Object { $_.id -eq $deptId1 }
Check '部门树包含 财务部 -> 会计组' ($null -ne $root -and @($root.children).Count -eq 1 -and $root.children[0].id -eq $deptId2) ($dt | ConvertTo-Json -Compress -Depth 6)
Check '部门层级 level 正确（1/2）' ($null -ne $root -and $root.level -eq 1 -and $root.children[0].level -eq 2) $(if ($root) { "$($root.level)/$($root.children[0].level)" } else { 'root missing' })

$selfParent = Api POST '/api/system/departments' (Body @{ id = $deptId1; deptCode = 'ZZE2E_D001'; deptName = '财务部'; parentId = $deptId2; leaderId = 0; sortOrder = 1; isActive = 1 }) $adminToken
Check '把部门挂到自己子部门下被拒绝' ($selfParent.code -ne '0000') ($selfParent | ConvertTo-Json -Compress)

$delBusy = Api DELETE "/api/system/departments/$deptId1" $null $adminToken
Check '有子部门时删除被拒绝' ($delBusy.code -ne '0000') ($delBusy | ConvertTo-Json -Compress)

Write-Host "`n=== 5. 部门菜单权限（会计组 = 员工所在部门） ==="
$sm = Api PUT "/api/system/departments/$deptId2/menus" ("{""menuIds"":[" + $idUser + "]}") $adminToken
Check '仅勾选 setting_user 保存成功' ($sm.code -eq '0000') ($sm | ConvertTo-Json -Compress)

$dm = Api GET "/api/system/departments/$deptId2/menus" $null $adminToken
$script:checkedCodes = @()
function WalkChecked($nodes) {
  foreach ($n in $nodes) {
    if ($n.checked) { $script:checkedCodes += $n.menuCode }
    if ($null -ne $n.children) { WalkChecked $n.children }
  }
}
WalkChecked $dm.data.list
Check '祖先目录 setting 被自动补齐' ($script:checkedCodes -contains 'setting') ($script:checkedCodes -join ',')
Check '部门实际授权为 setting + setting_user' ((@($script:checkedCodes) -join ',') -eq 'setting,setting_user') ($script:checkedCodes -join ',')
Check '未授权的 setting_employee 处于未勾选状态' (@($dm.data.list[0].children | Where-Object { $_.menuCode -eq 'setting_employee' -and $_.checked }).Count -eq 0)

$smBad = Api PUT "/api/system/departments/999999/menus" ("{""menuIds"":[" + $idUser + "]}") $adminToken
Check '给不存在的部门授权被拒绝' ($smBad.code -ne '0000') ($smBad | ConvertTo-Json -Compress)

# 记录当前设计语义：部门授权按节点独立维护，上级授权不会自动继承给子部门
$smParent = Api PUT "/api/system/departments/$deptId1/menus" ("{""menuIds"":[" + $idDept + "]}") $adminToken
Check '上级部门（财务部）可单独授权' ($smParent.code -eq '0000') ($smParent | ConvertTo-Json -Compress)
$dmChild = Api GET "/api/system/departments/$deptId2/menus" $null $adminToken
$script:checkedCodes = @()
WalkChecked $dmChild.data.list
Check '部门授权按节点独立（上级授权不下发给子部门）' ((@($script:checkedCodes) -join ',') -eq 'setting,setting_user') ($script:checkedCodes -join ',')

Write-Host "`n=== 6. 员工 CRUD 与部门受限授权 ==="
$e1 = Api POST '/api/system/employees' (Body @{ id = 0; empCode = 'ZZE2E_E001'; empName = '张三'; deptId = $deptId2; position = '会计'; phone = '13800000001'; email = 'z@x.com'; hireDate = '2024-01-15'; isActive = 1 }) $adminToken
Check '新增员工成功' ($e1.code -eq '0000') ($e1 | ConvertTo-Json -Compress)
$empId1 = $e1.data.id

$eBad = Api POST '/api/system/employees' (Body @{ id = 0; empCode = 'ZZE2E_E002'; empName = '李四'; deptId = 0; position = ''; phone = ''; email = ''; hireDate = ''; isActive = 1 }) $adminToken
Check '员工必须归属部门（不传部门被拒绝）' ($eBad.code -ne '0000') ($eBad | ConvertTo-Json -Compress)

$el = Api GET "/api/system/employees?deptId=$deptId1&page=1&pageSize=20" $null $adminToken
Check '员工列表分页接口可用（COUNT 修复）' ($el.code -eq '0000') ($el | ConvertTo-Json -Compress -Depth 4)
Check '按父部门筛选能查到子部门员工' (@($el.data.list).Count -eq 1 -and $el.data.list[0].id -eq $empId1 -and $el.data.total -eq 1) ($el | ConvertTo-Json -Compress -Depth 6)
Check '员工列表带出部门名称' ($el.data.list[0].deptName -eq '会计组') $el.data.list[0].deptName
Check '员工列表带出入职日期' ($el.data.list[0].hireDate -like '2024-01-15*') $el.data.list[0].hireDate

$el2 = Api GET '/api/system/employees?keyword=张三' $null $adminToken
Check '按姓名关键词搜索命中' (@($el2.data.list).Count -eq 1) ($el2 | ConvertTo-Json -Compress -Depth 4)

$elNo = Api GET '/api/system/employees?keyword=nobody-here' $null $adminToken
Check '无匹配时返回空列表且 total=0' (@($elNo.data.list).Count -eq 0 -and $elNo.data.total -eq 0) ($elNo | ConvertTo-Json -Compress -Depth 4)

$em = Api GET "/api/system/employees/$empId1/menus" $null $adminToken
$script:empMenuCodes = @()
function WalkAll($nodes) {
  foreach ($n in $nodes) { $script:empMenuCodes += $n.menuCode; if ($null -ne $n.children) { WalkAll $n.children } }
}
WalkAll $em.data.list
Check '员工可选菜单被限制在部门范围内（无 setting_employee）' (($script:empMenuCodes -contains 'setting_user') -and -not ($script:empMenuCodes -contains 'setting_employee')) ($script:empMenuCodes -join ',')

$oob = Api PUT "/api/system/employees/$empId1/menus" ("{""menuIds"":[" + $idEmp + "]}") $adminToken
Check '员工授权超出部门范围被拒绝' ($oob.code -ne '0000') ($oob | ConvertTo-Json -Compress)
Check '拒绝原因指向部门授权' ($oob.msg -like '*部门*') $oob.msg

$emOk = Api PUT "/api/system/employees/$empId1/menus" ("{""menuIds"":[" + $idUser + "]}") $adminToken
Check '员工授权在部门范围内可保存' ($emOk.code -eq '0000') ($emOk | ConvertTo-Json -Compress)

$el3 = Api GET '/api/system/employees?keyword=ZZE2E_E001' $null $adminToken
Check '员工列表统计已授权菜单数=2（含补齐的目录）' ($el3.data.list[0].menuCount -eq 2) $el3.data.list[0].menuCount

Write-Host "`n=== 7. 账号管理与绑定员工 ==="
$ul = Api GET '/api/system/users?page=1&pageSize=20' $null $adminToken
Check '账号列表分页接口可用（COUNT 修复）' ($ul.code -eq '0000') ($ul | ConvertTo-Json -Compress -Depth 4)

$u1 = Api POST '/api/system/users' (Body @{ id = 0; username = 'zze2e_user'; password = 'abc123456'; realName = '张三'; phone = '13800000001'; email = 'z@x.com'; role = 'staff'; departmentId = 0; status = 'active' }) $adminToken
Check '新增登录账号成功' ($u1.code -eq '0000') ($u1 | ConvertTo-Json -Compress)
$userId1 = $u1.data.id

$uDup = Api POST '/api/system/users' (Body @{ id = 0; username = 'zze2e_user'; password = 'abc123456'; realName = '重名'; phone = ''; email = ''; role = 'staff'; departmentId = 0; status = 'active' }) $adminToken
Check '重复用户名被拒绝' ($uDup.code -ne '0000') ($uDup | ConvertTo-Json -Compress)

$uNoPwd = Api POST '/api/system/users' (Body @{ id = 0; username = 'nopwd'; password = ''; realName = '无密码'; phone = ''; email = ''; role = 'staff'; departmentId = 0; status = 'active' }) $adminToken
Check '新增账号不填密码被拒绝' ($uNoPwd.code -ne '0000') ($uNoPwd | ConvertTo-Json -Compress)

$uBadRole = Api POST '/api/system/users' (Body @{ id = 0; username = 'badrole'; password = 'abc123456'; realName = ''; phone = ''; email = ''; role = 'god'; departmentId = 0; status = 'active' }) $adminToken
Check '非法角色被拒绝' ($uBadRole.code -ne '0000') ($uBadRole | ConvertTo-Json -Compress)

$eo = Api GET '/api/system/employeeOptions' $null $adminToken
Check '可绑定员工列表非空' (@($eo.data.list).Count -ge 1) ($eo | ConvertTo-Json -Compress)
Check '未绑定员工为可选状态' (($eo.data.list | Where-Object { $_.id -eq $empId1 }).disabled -eq $false)

$bind = Api PUT "/api/system/users/$userId1/bindEmployee" (Body @{ employeeId = $empId1 }) $adminToken
Check '账号绑定员工成功' ($bind.code -eq '0000') ($bind | ConvertTo-Json -Compress)

$eo2 = Api GET '/api/system/employeeOptions' $null $adminToken
$boundOpt = $eo2.data.list | Where-Object { $_.id -eq $empId1 }
Check '已绑定员工在选项中标记为不可选' ($boundOpt.disabled -eq $true) ($boundOpt | ConvertTo-Json -Compress)

$ul2 = Api GET '/api/system/users?keyword=zze2e_user' $null $adminToken
$row = $ul2.data.list | Where-Object { $_.id -eq $userId1 }
Check '账号列表带出部门与绑定员工' ($row.departmentName -eq '会计组' -and $row.employeeName -eq '张三') ($row | ConvertTo-Json -Compress)

Write-Host "`n=== 8. 新账号登录后的动态路由（部门∩员工） ==="
$login2 = Api POST '/auth/login' (Body @{ userName = 'zze2e_user'; password = 'abc123456' }) $null
Check '新账号可登录（密码哈希与登录逻辑一致）' ($login2.code -eq '0000') ($login2 | ConvertTo-Json -Compress)
$userToken = $login2.data.token

$ur2 = Api GET '/route/getUserRoutes' $null $userToken
$names2 = RouteNames $ur2.data.routes
Check '普通账号路由含 home/init/setting' (($names2 -contains 'home') -and ($names2 -contains 'init') -and ($names2 -contains 'setting')) ($names2 -join ',')
$settingNode2 = $ur2.data.routes | Where-Object { $_.name -eq 'setting' }
$childNames2 = RouteNames $settingNode2.children
Check '普通账号 setting 下只有 setting_user' ((@($childNames2).Count -eq 1) -and ($childNames2[0] -eq 'setting_user')) ($childNames2 -join ',')

$userInfo2 = Api GET '/auth/getUserInfo' $null $userToken
Check '普通账号角色为 R_STAFF' ($userInfo2.data.roles -contains 'R_STAFF') ($userInfo2 | ConvertTo-Json -Compress)

$ie3 = Api GET '/route/isRouteExist?routeName=setting_user' $null $userToken
Check '普通账号 isRouteExist(setting_user)=true' ($ie3.data -eq $true) ($ie3 | ConvertTo-Json -Compress)
$ie4 = Api GET '/route/isRouteExist?routeName=setting_employee' $null $userToken
Check '普通账号 isRouteExist(setting_employee)=false' ($ie4.data -eq $false) ($ie4 | ConvertTo-Json -Compress)

Write-Host "`n=== 9. 部门收权后员工权限自动收窄 ==="
$sm2 = Api PUT "/api/system/departments/$deptId2/menus" ("{""menuIds"":[" + $idDept + "]}") $adminToken
Check '员工所属部门改授权为 setting_department' ($sm2.code -eq '0000') ($sm2 | ConvertTo-Json -Compress)
$ur3 = Api GET '/route/getUserRoutes' $null $userToken
$names3 = RouteNames $ur3.data.routes
$settingNode3 = $ur3.data.routes | Where-Object { $_.name -eq 'setting' }
$childNames3 = RouteNames $settingNode3.children
Check '部门收权后员工 setting_user 立即失效' (-not ($childNames3 -contains 'setting_user')) ($childNames3 -join ',')
Check '交集只剩空目录 setting，已被剪除（不下发死链菜单）' (-not ($names3 -contains 'setting')) ($names3 -join ',')
Check '基础页面 home/init 仍然可达' (($names3 -contains 'home') -and ($names3 -contains 'init')) ($names3 -join ',')

Write-Host "`n=== 10. 重置密码、解绑与清理 ==="
$rp = Api PUT "/api/system/users/$userId1/password" (Body @{ password = 'newpass123' }) $adminToken
Check '重置密码成功' ($rp.code -eq '0000') ($rp | ConvertTo-Json -Compress)
$login3 = Api POST '/auth/login' (Body @{ userName = 'zze2e_user'; password = 'newpass123' }) $null
Check '重置后可用新密码登录' ($login3.code -eq '0000') ($login3 | ConvertTo-Json -Compress)
$login4 = Api POST '/auth/login' (Body @{ userName = 'zze2e_user'; password = 'abc123456' }) $null
Check '旧密码已失效' ($login4.code -ne '0000')

$shortPwd = Api PUT "/api/system/users/$userId1/password" (Body @{ password = '123' }) $adminToken
Check '过短密码被拒绝' ($shortPwd.code -ne '0000') ($shortPwd | ConvertTo-Json -Compress)

$delEmpBusy = Api DELETE "/api/system/employees/$empId1" $null $adminToken
Check '删除员工成功' ($delEmpBusy.code -eq '0000') ($delEmpBusy | ConvertTo-Json -Compress)
$ul3 = Api GET '/api/system/users?keyword=zze2e_user' $null $adminToken
$row2 = $ul3.data.list | Where-Object { $_.id -eq $userId1 }
Check '删除员工后账号自动解绑' (($null -eq $row2.employeeId) -or ($row2.employeeId -eq 0)) ($row2 | ConvertTo-Json -Compress)

$delUser = Api DELETE "/api/system/users/$userId1" $null $adminToken
Check '删除账号成功' ($delUser.code -eq '0000') ($delUser | ConvertTo-Json -Compress)

$delChild = Api DELETE "/api/system/departments/$deptId2" $null $adminToken
Check '删除空子部门成功' ($delChild.code -eq '0000') ($delChild | ConvertTo-Json -Compress)
$delParent = Api DELETE "/api/system/departments/$deptId1" $null $adminToken
Check '删除部门成功' ($delParent.code -eq '0000') ($delParent | ConvertTo-Json -Compress)

$dtEnd = Api GET '/api/system/departments' $null $adminToken
$leftDept = @($dtEnd.data.list | Where-Object { $_.deptCode -like 'ZZE2E*' })
Check '清理后无测试部门残留' ($leftDept.Count -eq 0) ($leftDept | ConvertTo-Json -Compress -Depth 4)
$elEnd = Api GET '/api/system/employees?page=1&pageSize=200' $null $adminToken
$leftEmp = @($elEnd.data.list | Where-Object { $_.empCode -like 'ZZE2E*' })
Check '清理后无测试员工残留' ($leftEmp.Count -eq 0) ($leftEmp | ConvertTo-Json -Compress -Depth 4)
$otherDept = @($dtEnd.data.list | Where-Object { $_.deptCode -notlike 'ZZE2E*' })
$otherEmp = @($elEnd.data.list | Where-Object { $_.empCode -notlike 'ZZE2E*' })
Write-Host ("  （库中其余业务数据：部门 " + $otherDept.Count + " 个，员工 " + $otherEmp.Count + " 条，未被本测试触碰）")

Write-Host "`n================================"
Write-Host ("PASS = " + $script:pass + "   FAIL = " + $script:fail)
Write-Host "================================"
if ($script:fail -gt 0) { exit 1 } else { exit 0 }
