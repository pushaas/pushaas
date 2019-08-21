import React from 'react'

// import configService from 'services/configService'

import { useStyles } from 'components/App/Private/styles'
import SetTitleContext from 'components/contexts/SetTitleContext'
import TitleContext from 'components/contexts/TitleContext'

import Header from './Header'
import Menu from './Menu'
import Main from './Main'

const Private = (props) => {
  const classes = useStyles()
  const [open, setOpen] = React.useState(true)
  const [title, setTitle] = React.useState('')
  const handleDrawerOpen = () => { setOpen(true) }
  const handleDrawerClose = () => { setOpen(false) }

  return (
    <div className={classes.root}>
      <SetTitleContext.Provider value={setTitle}>
        <TitleContext.Provider value={title}>
          <Header open={open} handleDrawerOpen={handleDrawerOpen} />
          <Menu open={open} handleDrawerClose={handleDrawerClose} />
          <Main />
        </TitleContext.Provider>
      </SetTitleContext.Provider>
    </div>
  )
}

export default  Private
