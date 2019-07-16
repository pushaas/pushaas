import React, { useContext, useEffect, useState } from 'react'
import { Redirect } from 'react-router-dom'
import clsx from 'clsx'

import Grid from '@material-ui/core/Grid'
import Paper from '@material-ui/core/Paper'

import { privateResourcesPath } from 'navigation'

import resourcesService from 'services/resourcesService'

import SetTitleContext from 'components/contexts/SetTitleContext'
import { useStyles } from 'components/App/Private/styles'

const Channels = (props) => {
  const id = props.match.params.id
  const classes = useStyles()
  const [didLoad, setDidLoad] = useState(false)
  const [resources, setResources] = useState([])
  const setTitle = useContext(SetTitleContext)

  const findSelectedChannelById = () => {
    if (id && resources.length) {
      return resources.find(c => c.id === id)
    }
  }
  const selectedChannel = id ? findSelectedChannelById() : undefined

  useEffect(() => {
    setTitle('Persistent Channels')
  }, [setTitle])

  useEffect(() => {
    resourcesService.getResources()
      .then((data) => {
        setResources(data)
        setDidLoad(true)
      })
  }, [])

  const resourcesMinHeightPaper = clsx(classes.paper, classes.resourcesMinHeightPaper)

  if (didLoad && id && !selectedChannel) {
    return (
      <Redirect to={privateResourcesPath} />
    )
  }

  return (
    <Grid container spacing={3}>
      <Grid item xs={6}>
        Resources
      </Grid>
      <Grid item xs={6}>
      </Grid>
    </Grid>
  )
}

export default Channels
