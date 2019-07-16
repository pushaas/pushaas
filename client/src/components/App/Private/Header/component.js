import React, { useContext } from 'react'
import clsx from 'clsx'
import AppBar from '@material-ui/core/AppBar'
import Toolbar from '@material-ui/core/Toolbar'
import Typography from '@material-ui/core/Typography'
import IconButton from '@material-ui/core/IconButton'
import MenuIcon from '@material-ui/icons/Menu'
import ExitToAppIcon from '@material-ui/icons/ExitToApp'

import credentialsService from 'services/credentialsService'

import { useStyles } from 'components/App/Private/styles'
import SetUserContext from 'components/contexts/SetUserContext'
import TitleContext from 'components/contexts/TitleContext'

const Header = ({ open, handleDrawerOpen }) => {
  const classes = useStyles()
  const setUser = useContext(SetUserContext)
  const title = useContext(TitleContext)

  const logout = () => {
    credentialsService.clearCredentials()
    setUser(null)
  }

  return (
    <AppBar position="absolute" className={clsx(classes.appBar, open && classes.appBarShift)}>
      <Toolbar className={classes.toolbar}>
        <IconButton
          edge="start"
          color="inherit"
          aria-label="Open drawer"
          onClick={handleDrawerOpen}
          className={clsx(classes.menuButton, open && classes.menuButtonHidden)}
        >
          <MenuIcon />
        </IconButton>
        <Typography component="h1" variant="h6" color="inherit" noWrap className={classes.title}>
          {title}
        </Typography>
        <IconButton color="inherit" onClick={logout}>
          <ExitToAppIcon />
        </IconButton>
      </Toolbar>
    </AppBar>
  )
}

export default Header
