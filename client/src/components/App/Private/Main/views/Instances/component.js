import React, { useContext, useEffect, useState } from 'react'
import { Redirect } from 'react-router-dom'
import clsx from 'clsx'

import Grid from '@material-ui/core/Grid'
import Paper from '@material-ui/core/Paper'

import { privateInstancesPath } from 'navigation'

import instancesService from 'services/instancesService'

import SetTitleContext from 'components/contexts/SetTitleContext'
import { useStyles } from 'components/App/Private/styles'
import InstanceList from './InstanceList'

const Instances = (props) => {
  const id = props.match.params.id
  const classes = useStyles()
  const [didLoad, setDidLoad] = useState(false)
  const [instances, setInstances] = useState([])
  const setTitle = useContext(SetTitleContext)

  const findSelectedInstanceById = () => {
    if (id && instances.length) {
      return instances.find(c => c.id === id)
    }
  }
  const selectedInstance = id ? findSelectedInstanceById() : undefined

  useEffect(() => {
    setTitle('Persistent Instances')
  }, [setTitle])

  useEffect(() => {
    instancesService.getInstances()
      .then((data) => {
        setInstances(data)
        setDidLoad(true)
      })
  }, [])

  const instancesMinHeightPaper = clsx(classes.paper, classes.instancesMinHeightPaper)

  if (didLoad && id && !selectedInstance) {
    return (
      <Redirect to={privateInstancesPath} />
    )
  }

  return (
    <Grid container>
      <Grid item xs={12}>
        <Paper className={instancesMinHeightPaper}>
          <InstanceList instances={instances} />
        </Paper>
      </Grid>
    </Grid>
  )
}

export default Instances
