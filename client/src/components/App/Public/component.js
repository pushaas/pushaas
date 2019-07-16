import React from 'react'
import { Redirect, Route, Switch } from 'react-router-dom'

import { publicLoginPath } from 'navigation'

import Login from './Login'

const Public = (props) => (
  <Switch>
    <Route path={publicLoginPath} component={Login} />
    <Redirect to={publicLoginPath} />
  </Switch>
)

export default  Public
