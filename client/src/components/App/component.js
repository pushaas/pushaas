import React, { useEffect, useState } from 'react'
import { BrowserRouter as Router } from 'react-router-dom'
import CssBaseline from '@material-ui/core/CssBaseline'
import CircularProgress from '@material-ui/core/CircularProgress'
import { ToastContainer } from 'react-toastify'

import { useStyles } from 'components/App/styles'

import authService from 'services/authService'
import credentialsService from 'services/credentialsService'

import { routerBaseName } from 'navigation'

import SetUserContext from 'components/contexts/SetUserContext'
import Private from './Private'
import Public from './Public'

const STATUS_LOADING = 'STATUS_LOADING'
const STATUS_LOADED = 'STATUS_LOADED'

const App = () => {
  const classes = useStyles()

  const [user, setUser] = useState(null)
  const [status, setStatus] = useState(STATUS_LOADING)

  useEffect(() => {
    const credentials = credentialsService.getCredentials()
    if (!credentials) {
      setStatus(STATUS_LOADED)
      return
    }

    const { username } = credentials
    authService.checkAuth(credentials)
      .then(() => setUser({ username }))
      .finally(() => setStatus(STATUS_LOADED))
  }, [setStatus, setUser])

  const renderLoading = () => <div className={classes.loader}><CircularProgress /></div>
  const renderLoaded = () => user ? <Private /> : <Public />

  return (
    <React.Fragment>
      <CssBaseline />
      <ToastContainer />

      <Router basename={routerBaseName}>
        <SetUserContext.Provider value={setUser}>
          {status === STATUS_LOADING ? renderLoading() : renderLoaded()}
        </SetUserContext.Provider>
      </Router>
    </React.Fragment>
  )
}

export default App
