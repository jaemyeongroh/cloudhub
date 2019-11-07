import React, {PureComponent} from 'react'
import {connect} from 'react-redux'
import {bindActionCreators} from 'redux'

// Components
import Threesizer from 'src/shared/components/threesizer/Threesizer'
import AgentTable from 'src/agent_admin/components/AgentTable'
import FancyScrollbar from 'src/shared/components/FancyScrollbar'

import {ErrorHandling} from 'src/shared/decorators/errors'

//const
import {HANDLE_HORIZONTAL, HANDLE_VERTICAL} from 'src/shared/constants'

interface State {
  minions: Readonly<[]>
  proportions: number[]
}

@ErrorHandling
class AgentConfiguration extends PureComponent<State> {
  constructor(props) {
    super(props)
    this.state = {
      measurements: [],
      minionLog: 'not load log',
      horizontalProportions: [0.43, 0.57],
      verticalProportions: [0.43, 0.57],
    }

    this.measurementsTemp = this.measurementsTemp.bind(this)
  }

  public onClickTableRowCall() {
    return console.log('row Called', this)
  }

  public onClickActionCall() {
    return console.log('action Called', this)
  }

  public onClickSaveCall() {
    return console.log('Save Called', this)
  }

  public onClickTestCall() {
    return console.log('Test Called', this)
  }

  public onClickApplyCall() {
    return console.log('Apply Called', this)
  }

  public componentDidMount() {
    this.setState({
      measurements: [
        'tomcat',
        'redis',
        'mySQL',
        'mongoDB',
        'docker',
        'cpu',
        'memory',
      ],
    })
  }

  render() {
    return (
      <div className="panel panel-solid">
        <Threesizer
          orientation={HANDLE_HORIZONTAL}
          divisions={this.horizontalDivisions}
          onResize={this.horizontalHandleResize}
        />
      </div>
    )
  }

  private horizontalHandleResize = (horizontalProportions: number[]) => {
    this.setState({horizontalProportions})
  }

  private verticalHandleResize = (verticalProportions: number[]) => {
    this.setState({verticalProportions})
  }

  private renderAgentPageTop = () => {
    const {currentUrl, minions} = this.props
    return (
      <AgentTable
        currentUrl={currentUrl}
        minions={minions}
        onClickTableRow={this.onClickTableRowCall}
        onClickAction={this.onClickActionCall}
      />
    )
  }

  private renderAgentPageBottom = () => {
    const {minionLog} = this.state
    return (
      <Threesizer
        orientation={HANDLE_VERTICAL}
        divisions={this.verticalDivisions}
        onResize={this.verticalHandleResize}
      />
    )
  }

  private measurementsTemp() {
    const {measurements} = this.state
    return (
      <div className="panel">
        <div className="panel-heading">
          <h2
            className="panel-title"
            style={{
              width: '100%',
            }}
          >
            measurements
            <div
              style={{
                color: '#f58220',
                fontSize: '12px',
                background: '#232323',
                padding: '10px',
                margin: '5px 0px',
                width: '100%',
              }}
            >
              host1-minion1-192.168.0.1
            </div>
          </h2>
        </div>
        <div className="panel-body">
          <FancyScrollbar>
            <div className="query-builder--list">
              {measurements.map((v, i) => {
                return (
                  <div
                    className="query-builder--list-item"
                    data-key={i}
                  >{`${v}`}</div>
                )
              })}
            </div>
          </FancyScrollbar>
        </div>
      </div>
    )
  }

  private collectorConfigTemp() {
    return (
      <div className="panel">
        <div className="panel-heading">
          <h2 className="panel-title">collector.conf</h2>
          <div>
            <button
              className="btn btn-inline_block btn-default"
              onClick={this.onClickSaveCall}
            >
              SAVE
            </button>
            <button
              className="btn btn-inline_block btn-default"
              style={{
                marginLeft: '5px',
              }}
              onClick={this.onClickTestCall}
            >
              TEST
            </button>
            <button
              className="btn btn-inline_block btn-default"
              style={{
                marginLeft: '5px',
              }}
              onClick={this.onClickApplyCall}
            >
              APPLY
            </button>
          </div>
        </div>

        <div className="panel-body">
          <div
            style={{
              width: '100%',
              height: '100%',
              overflow: 'hidden',
              position: 'relative',
            }}
          >
            <FancyScrollbar>
              <textarea
                style={{
                  resize: 'none',
                  width: '100%',
                  height: '100%',
                  backgroundColor: '#232323',
                  color: '#fff',
                  borderColor: 'transparent',
                  padding: '15px',
                }}
              >
                {`
#[global_tags]
#  dc = "denver-1"

#[agent]
# interval = "10s"

# OUTPUTS
#[[outputs.influxdb]]
#  url = "http://192.168.59.103:8086" # required.
#  database = "telegraf" # required.
# precision = "s"


#[global_tags]
#  dc = "denver-1"

#[agent]
# interval = "10s"

# OUTPUTS
#[[outputs.influxdb]]
#  url = "http://192.168.59.103:8086" # required.
#  database = "telegraf" # required.
# precision = "s"


#[global_tags]
#  dc = "denver-1"

#[agent]
# interval = "10s"

# OUTPUTS
#[[outputs.influxdb]]
#  url = "http://192.168.59.103:8086" # required.
#  database = "telegraf" # required.
# precision = "s"


#[global_tags]
#  dc = "denver-1"

#[agent]
# interval = "10s"

# OUTPUTS
#[[outputs.influxdb]]
#  url = "http://192.168.59.103:8086" # required.
#  database = "telegraf" # required.
# precision = "s"


#[global_tags]
#  dc = "denver-1"

#[agent]
# interval = "10s"

# OUTPUTS
#[[outputs.influxdb]]
#  url = "http://192.168.59.103:8086" # required.
#  database = "telegraf" # required.
# precision = "s"
              `}
              </textarea>
            </FancyScrollbar>
          </div>
        </div>
      </div>
    )
  }

  private get horizontalDivisions() {
    const {horizontalProportions} = this.state
    const [topSize, bottomSize] = horizontalProportions

    return [
      {
        name: '',
        handleDisplay: 'none',
        headerButtons: [],
        menuOptions: [],
        render: this.renderAgentPageTop,
        headerOrientation: HANDLE_HORIZONTAL,
        size: topSize,
      },
      {
        name: '',
        handlePixels: 8,
        headerButtons: [],
        menuOptions: [],
        render: this.renderAgentPageBottom,
        headerOrientation: HANDLE_HORIZONTAL,
        size: bottomSize,
      },
    ]
  }

  private get verticalDivisions() {
    const {verticalProportions} = this.state
    const [rightSize, leftSize] = verticalProportions

    return [
      {
        name: '',
        handleDisplay: 'none',
        headerButtons: [],
        menuOptions: [],
        render: this.measurementsTemp,
        headerOrientation: HANDLE_VERTICAL,
        size: rightSize,
      },
      {
        name: '',
        handlePixels: 8,
        headerButtons: [],
        menuOptions: [],
        render: this.collectorConfigTemp.bind(this),
        headerOrientation: HANDLE_VERTICAL,
        size: leftSize,
      },
    ]
  }
}

export default AgentConfiguration
