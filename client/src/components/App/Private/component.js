import React, { useEffect } from 'react'

// import configService from 'services/configService'

import { useStyles } from 'components/App/Private/styles'
import ConfigContext from 'components/contexts/ConfigContext'
import SetTitleContext from 'components/contexts/SetTitleContext'
import TitleContext from 'components/contexts/TitleContext'

import Header from './Header'
import Menu from './Menu'
import Main from './Main'

const Private = (props) => {
  const classes = useStyles()
  const [config, setConfig] = React.useState(null)
  const [open, setOpen] = React.useState(true)
  const [title, setTitle] = React.useState('')
  const handleDrawerOpen = () => { setOpen(true) }
  const handleDrawerClose = () => { setOpen(false) }

  // TODO
  // useEffect(() => {
  //   configService.getConfig()
  //     .then((data) => setConfig(data))
  // }, [])

  return (
    <div className={classes.root}>
      <ConfigContext.Provider value={config}>
        <SetTitleContext.Provider value={setTitle}>
          <TitleContext.Provider value={title}>
            <Header open={open} handleDrawerOpen={handleDrawerOpen} />
            <Menu open={open} handleDrawerClose={handleDrawerClose} />
            <Main />
          </TitleContext.Provider>
        </SetTitleContext.Provider>
      </ConfigContext.Provider>
    </div>
  )
}

export default  Private
