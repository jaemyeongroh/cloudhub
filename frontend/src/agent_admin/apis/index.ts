import axios from 'axios'
import {Minion} from 'src/agent_admin/type'

interface MinionsObject {
  [x: string]: Minion
}

interface Params {
  client?: string
  fun?: string
  arg?: string[] | string
  tgt_type?: string
  tgt?: string[] | string
  match?: string
  include_rejected?: string
  include_denied?: string
  include_accepted?: string
  show_ip?: string
  kwarg?: {
    name?: string
    path?: string
    dest?: string
    makedirs?: string
    fun?: string
    cmd?: string
    args?: string[] | string
  }
  username?: string
  password?: string
  eauth?: string
}

const EmptyMionin: Minion = {
  host: '',
  ip: '',
  os: '',
  osVersion: '',
  status: '',
  isCheck: false,
}

const apiRequest = async (
  pMethod: string,
  pRoute: string,
  pParams: Params
): Promise<MinionsObject | any> => {
  const dParams = {}
  const saltMasterUrl = window.localStorage.getItem('salt-master-url')
  const url = saltMasterUrl + pRoute
  const token = window.localStorage.getItem('salt-master-token')
  const headers = {
    Accept: 'application/json',
    'X-Auth-Token': token !== null ? token : '',
    'Content-type': 'application/json',
  }

  Object.assign(dParams, pParams)

  const param = JSON.stringify(dParams)

  return axios({
    method: pMethod,
    url: url,
    headers: headers,
    data: param,
  })
    .then(response => {
      return response
    })
    .catch(error => {
      return error
    })
}

export function getSaltToken(
  pUserName: string,
  pPassWord: string,
  pEauth: string = 'pam'
) {
  const params = {
    username: pUserName,
    password: pPassWord,
    eauth: pEauth,
  }

  // store it as the default login method
  window.localStorage.setItem('eauth', pEauth)

  return apiRequest('POST', '/login', params)
}

export const getMinionKeyListAllAsync = async (): Promise<MinionsObject> => {
  const minions: MinionsObject = {}
  const info = await Promise.all([
    getWheelKeyListAll(),
    getRunnerManageAllowed(),
    getLocalGrainsItems(''),
  ])

  const info2 = await Promise.all([
    getLocalServiceEnabledTelegraf(info[0].data.return[0].data.return.minions),
    getLocalServiceStatusTelegraf(info[0].data.return[0].data.return.minions),
  ])

  const keyList = info[0].data.return[0].data.return.minions
  const ipList = info[1].data.return[0]
  const osList = info[2].data.return[0]

  const installList = info2[0].data.return[0]
  const statusList = info2[1].data.return[0]

  for (const k of keyList)
    minions[k] = {
      host: k,
      status: 'Accept',
      isCheck: false,
      ip: ipList[k],
      os: osList[k].os,
      osVersion: osList[k].osrelease,
      isInstall: installList[k] != true ? false : installList[k],
      isRunning: statusList[k],
    }

  return minions
}

export const getMinionKeyListAll = async (): Promise<MinionsObject> => {
  const minions: MinionsObject = {}
  const wheelKeyListAllPromise = getWheelKeyListAll()

  return wheelKeyListAllPromise.then(pWheelKeyListAllData => {
    for (const k of pWheelKeyListAllData.data.return[0].data.return.minions)
      minions[k] = {
        ...EmptyMionin,
        host: k,
        status: 'Accept',
      }

    for (const k of pWheelKeyListAllData.data.return[0].data.return.minions_pre)
      minions[k] = {
        ...EmptyMionin,
        host: k,
        status: 'UnAccept',
      }

    for (const k of pWheelKeyListAllData.data.return[0].data.return
      .minions_rejected)
      minions[k] = {
        ...EmptyMionin,
        host: k,
        status: 'ReJect',
      }

    return minions
  })
}

export const getMinionAcceptKeyListAll = async (): Promise<MinionsObject> => {
  const minions: MinionsObject = {}
  const wheelKeyListAllPromise = getWheelKeyListAll()

  return wheelKeyListAllPromise.then(pWheelKeyListAllData => {
    for (const k of pWheelKeyListAllData.data.return[0].data.return.minions)
      minions[k] = {
        ...EmptyMionin,
        host: k,
        status: 'Accept',
      }

    return minions
  })
}

export const getMinionsIP = async (
  minions: MinionsObject
): Promise<MinionsObject> => {
  const newMinions = {...minions}

  const getRunnerManageAllowedPromise = getRunnerManageAllowed()
  return getRunnerManageAllowedPromise.then(pRunnerManageAllowedData => {
    Object.keys(pRunnerManageAllowedData.data.return[0]).forEach(function(k) {
      newMinions[k] = {
        host: k,
        status: newMinions[k].status,
        isCheck: newMinions[k].isCheck,
        ip: pRunnerManageAllowedData.data.return[0][k],
      }
    })
    return newMinions
  })
}

export const getMinionsOS = async (
  minions: MinionsObject
): Promise<MinionsObject> => {
  const newMinions = {...minions}
  const getLocalGrainsItemsPromise = getLocalGrainsItems('')

  return getLocalGrainsItemsPromise.then(pLocalGrainsItemsData => {
    Object.keys(pLocalGrainsItemsData.data.return[0]).forEach(function(k) {
      if (newMinions.hasOwnProperty(k)) {
        newMinions[k] = {
          host: k,
          status: newMinions[k].status,
          isCheck: newMinions[k].isCheck,
          ip: newMinions[k].ip,
          os: pLocalGrainsItemsData.data.return[0][k].os,
          osVersion: pLocalGrainsItemsData.data.return[0][k].osrelease,
        }
      }
    })
    return newMinions
  })
}

export const getTelegrafInstalled = async (
  minions: MinionsObject
): Promise<MinionsObject> => {
  const newMinions = {...minions}
  const getLocalServiceEnabledTelegrafPromise = getLocalServiceEnabledTelegraf(
    Object.keys(newMinions)
  )

  return getLocalServiceEnabledTelegrafPromise.then(
    pLocalServiceEnabledTelegrafData => {
      Object.keys(pLocalServiceEnabledTelegrafData.data.return[0]).forEach(
        function(k) {
          if (newMinions.hasOwnProperty(k)) {
            newMinions[k] = {
              host: k,
              status: newMinions[k].status,
              isCheck: newMinions[k].isCheck,
              ip: newMinions[k].ip,
              os: newMinions[k].os,
              osVersion: newMinions[k].osVersion,
              isInstall:
                pLocalServiceEnabledTelegrafData.data.return[0][k] != true
                  ? false
                  : pLocalServiceEnabledTelegrafData.data.return[0][k],
            }
          }
        }
      )
      return newMinions
    }
  )
}

export const getTelegrafServiceStatus = async (
  minions: MinionsObject
): Promise<MinionsObject> => {
  const newMinions = {...minions}
  const getLocalServiceStatusTelegrafPromise = getLocalServiceStatusTelegraf(
    Object.keys(newMinions)
  )

  return getLocalServiceStatusTelegrafPromise.then(
    pLocalServiceStatusTelegrafData => {
      Object.keys(pLocalServiceStatusTelegrafData.data.return[0]).forEach(
        function(k) {
          if (newMinions.hasOwnProperty(k)) {
            newMinions[k] = {
              host: k,
              status: newMinions[k].status,
              isCheck: newMinions[k].isCheck,
              ip: newMinions[k].ip,
              os: newMinions[k].os,
              osVersion: newMinions[k].osVersion,
              isInstall: newMinions[k].isInstall,
              isRunning: pLocalServiceStatusTelegrafData.data.return[0][k],
            }
          }
        }
      )
      return newMinions
    }
  )
}

export function getLocalGrainsItem(pMinionId: string) {
  const params = {
    client: 'local',
    tgt: pMinionId,
    fun: 'grains.item',
    arg: [
      'saltversion',
      'master',
      'os_family',
      'os',
      'osrelease',
      'kernel',
      'kernelrelease',
      'kernelversion',
      'virtual',
      'cpuarch',
      'cpu_model',
      'localhost',
      'ip_interfaces',
      'ip6_interfaces',
      'ip4_gw',
      'ip6_gw',
      'dns:nameservers',
      'locale_info',
      'cpu_model',
      'biosversion',
      'mem_total',
      'swap_total',
      'gpus',
      'selinux',
      'path',
    ],
  }

  return apiRequest('POST', '/', params)
}

export function runAcceptKey(pMinionId: string) {
  const params = {
    client: 'wheel',
    fun: 'key.accept',
    match: pMinionId,
    include_rejected: 'true',
    include_denied: 'true',
  }

  return apiRequest('POST', '/', params)
}

export function runRejectKey(pMinionId: string) {
  const params = {
    client: 'wheel',
    fun: 'key.reject',
    match: pMinionId,
    include_accepted: 'true',
  }

  return apiRequest('POST', '/', params)
}

export function runDeleteKey(pMinionId: string) {
  const params = {
    client: 'wheel',
    fun: 'key.delete',
    match: pMinionId,
  }

  return apiRequest('POST', '/', params)
}

export function getWheelKeyListAll() {
  const params = {
    client: 'wheel',
    fun: 'key.list_all',
  }

  return apiRequest('POST', '/', params)
}

export function getRunnerManageAllowed() {
  const params = {
    client: 'runner',
    fun: 'manage.allowed',
    show_ip: 'true',
  }

  return apiRequest('POST', '/', params)
}

export function getLocalServiceEnabledTelegraf(pMinionId: string[]) {
  const params: Params = {
    client: 'local',
    fun: 'service.enabled',
    arg: 'telegraf',
    tgt_type: '',
    tgt: '',
  }
  if (pMinionId) {
    params.tgt_type = 'list'
    params.tgt = pMinionId
  } else {
    params.tgt_type = 'glob'
    params.tgt = '*'
  }
  return apiRequest('POST', '/', params)
}

export function getLocalServiceStatusTelegraf(pMinionId: string[]) {
  const params: Params = {
    client: 'local',
    fun: 'service.status',
    arg: 'telegraf',
    tgt_type: '',
    tgt: '',
  }
  if (pMinionId) {
    params.tgt_type = 'list'
    params.tgt = pMinionId
  } else {
    params.tgt_type = 'glob'
    params.tgt = '*'
  }
  return apiRequest('POST', '/', params)
}

export function runLocalServiceStartTelegraf(pMinionId: string | string[]) {
  const params: Params = {
    client: 'local',
    fun: 'service.start',
    arg: 'telegraf',
    tgt_type: '',
    tgt: '',
  }
  if (pMinionId) {
    params.tgt_type = 'list'
    params.tgt = pMinionId
  } else {
    params.tgt_type = 'glob'
    params.tgt = '*'
  }
  return apiRequest('POST', '/', params)
}

export function runLocalServiceStopTelegraf(pMinionId: string | string[]) {
  const params: Params = {
    client: 'local',
    fun: 'service.stop',
    arg: 'telegraf',
    tgt_type: '',
    tgt: '',
  }
  if (pMinionId) {
    params.tgt_type = 'list'
    params.tgt = pMinionId
  } else {
    params.tgt_type = 'glob'
    params.tgt = '*'
  }
  return apiRequest('POST', '/', params)
}

export function runLocalServiceReStartTelegraf(pMinionId: string) {
  const params: Params = {
    client: 'local',
    fun: 'service.restart',
    arg: 'telegraf',
    tgt_type: '',
    tgt: '',
  }
  if (pMinionId) {
    params.tgt_type = 'list'
    params.tgt = pMinionId
  } else {
    params.tgt_type = 'glob'
    params.tgt = '*'
  }
  return apiRequest('POST', '/', params)
}

export function runLocalCpGetDirTelegraf(pMinionId: string[]) {
  const params: Params = {
    client: 'local',
    fun: 'cp.get_dir',
    kwarg: {
      path: 'salt://telegraf',
      dest: '/srv/salt/prod',
      makedirs: 'true',
    },
    tgt_type: '',
    tgt: '',
  }
  if (pMinionId) {
    params.tgt_type = 'list'
    params.tgt = pMinionId
  } else {
    params.tgt_type = 'glob'
    params.tgt = '*'
  }
  return apiRequest('POST', '/', params)
}

export function runLocalPkgInstallTelegraf(pMinionId: string[]) {
  const params: Params = {
    client: 'local',
    fun: 'pkg.install',
    kwarg: {
      name: '//srv/salt/prod/telegraf/telegraf-1.12.4-1.x86_64.rpm',
    },
    tgt_type: '',
    tgt: '',
  }
  if (pMinionId) {
    params.tgt_type = 'list'
    params.tgt = pMinionId
  } else {
    params.tgt_type = 'glob'
    params.tgt = '*'
  }
  return apiRequest('POST', '/', params)
}

export function getLocalGrainsItems(pMinionId: string) {
  const params: Params = {
    client: 'local',
    fun: 'grains.items',
    tgt_type: '',
    tgt: '',
  }
  if (pMinionId) {
    params.tgt_type = 'list'
    params.tgt = pMinionId
  } else {
    params.tgt_type = 'glob'
    params.tgt = '*'
  }
  return apiRequest('POST', '/', params)
}

export function getLocalFileRead(pMinionId: string) {
  const params: Params = {
    client: 'local',
    fun: 'file.read',
    tgt_type: '',
    tgt: '',
    kwarg: {
      path: '/etc/telegraf/telegraf.conf',
    },
  }
  if (pMinionId) {
    params.tgt_type = 'list'
    params.tgt = pMinionId
  } else {
    params.tgt_type = 'glob'
    params.tgt = '*'
  }
  return apiRequest('POST', '/', params)
}

export function getLocalFileWrite(pMinionId: string, pScript: string) {
  const params: Params = {
    client: 'local',
    fun: 'file.write',
    tgt_type: '',
    tgt: '',
    kwarg: {
      path: '/etc/telegraf/telegraf.conf',
      args: [pScript],
    },
  }
  if (pMinionId) {
    params.tgt_type = 'list'
    params.tgt = pMinionId
  } else {
    params.tgt_type = 'glob'
    params.tgt = '*'
  }
  return apiRequest('POST', '/', params)
}

export function getLocalServiceGetRunning(pMinionId: string) {
  const params: Params = {
    client: 'local',
    fun: 'service.get_running',
    tgt_type: '',
    tgt: '',
  }
  if (pMinionId) {
    params.tgt_type = 'list'
    params.tgt = pMinionId
  } else {
    params.tgt_type = 'glob'
    params.tgt = '*'
  }
  return apiRequest('POST', '/', params)
}

export function getRunnerSaltCmdTelegraf(pMeasurements: string) {
  const params = {
    client: 'runner',
    fun: 'salt.cmd',
    kwarg: {
      fun: 'cmd.run',
      cmd: 'telegraf --usage ' + pMeasurements,
    },
  }

  return apiRequest('POST', '/', params)
}