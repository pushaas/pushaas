import React from 'react'
import { Redirect, Route, Switch } from 'react-router-dom'

import Container from '@material-ui/core/Container'

import { privateHomePath, privateInstancesPathWithOptionalParam } from 'navigation'
import { useStyles } from 'components/App/Private/styles'

import Dashboard from './views/Dashboard'
import Instances from './views/Instances'

const Main = (props) => {
  const classes = useStyles()
  return (
    <main className={classes.content}>
      <div className={classes.appBarSpacer} />
      <Container maxWidth="lg" className={classes.container}>
        <Switch>
          <Route path={privateHomePath} exact component={Dashboard} />
          <Route path={privateInstancesPathWithOptionalParam} component={Instances} />
          <Redirect to={privateHomePath} />
        </Switch>
      </Container>
    </main>
  )
}

export default Main
