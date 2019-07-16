import React from 'react'
import { Redirect, Route, Switch } from 'react-router-dom'

import Container from '@material-ui/core/Container'

import { privateHomePath, privateResourcesPathWithOptionalParam } from 'navigation'
import { useStyles } from 'components/App/Private/styles'

import Dashboard from './views/Dashboard'
import Resources from './views/Resources'

const Main = (props) => {
  const classes = useStyles()
  return (
    <main className={classes.content}>
      <div className={classes.appBarSpacer} />
      <Container maxWidth="lg" className={classes.container}>
        <Switch>
          <Route path={privateHomePath} exact component={Dashboard} />
          <Route path={privateResourcesPathWithOptionalParam} component={Resources} />
          <Redirect to={privateHomePath} />
        </Switch>
      </Container>
    </main>
  )
}

export default Main
