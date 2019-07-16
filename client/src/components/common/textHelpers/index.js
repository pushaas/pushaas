import React from 'react'
import Typography from '@material-ui/core/Typography'

export const renderSubtitle1 = (text) => (<Typography component="p" variant="subtitle1" color="primary" gutterBottom>{text}</Typography>)
export const renderSubtitle2 = (text) => (<Typography component="p" variant="subtitle2" color="primary" gutterBottom>{text}</Typography>)
export const renderBody2 = (text) => (<Typography component="span" display="inline" variant="body2">{text}</Typography>)

export const renderTextItem = (label, value) => (
  <div>
    <Typography component="span" display="inline" variant="overline" color="primary" >{label} </Typography>
    {renderBody2(value)}
  </div>
)
